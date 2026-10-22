package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/stretchr/testify/require"
)

func TestServer_CheckOption_TripPublication(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		options            gen.ContractOptions
		disablePublication bool
		wantPublication    bool
		wantCreation       bool
	}{
		{
			name:    "creation permission does not grant publication",
			options: gen.ContractOptions{{Code: gen.TripCreation}}, wantCreation: true,
		},
		{
			name:    "publication permission does not grant creation",
			options: gen.ContractOptions{{Code: gen.TripPublication}}, wantPublication: true,
		},
		{
			name:            "both permissions are enabled",
			options:         gen.ContractOptions{{Code: gen.TripCreation}, {Code: gen.TripPublication}},
			wantPublication: true, wantCreation: true,
		},
		{
			name:               "publication is disabled independently",
			options:            gen.ContractOptions{{Code: gen.TripCreation}, {Code: gen.TripPublication}},
			disablePublication: true, wantCreation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			companyID := "company-" + uuid.NewString()
			created := sendCreateContractRequest(t, http.StatusCreated, gen.CreateContractRequest{
				CompanyID: companyID, StartsAt: time.Now().UTC().Add(-time.Hour), Options: tt.options,
			})
			require.Empty(t, created.Errors)
			require.NotNil(t, created.Data)
			require.ElementsMatch(t, tt.options, created.Data.Options)
			activated := sendActivateContractRequest(t, http.StatusCreated, gen.ActivateContractRequest{ID: created.Data.ID})
			require.Empty(t, activated.Errors)
			if tt.disablePublication {
				tag, err := testPool.Exec(t.Context(),
					`UPDATE contract_options SET enabled = false WHERE contract_id = $1 AND code = $2`,
					created.Data.ID, string(gen.TripPublication))
				require.NoError(t, err)
				require.EqualValues(t, 1, tag.RowsAffected())
			}

			publication := sendCheckOptionRequest(t, http.StatusOK, gen.CheckOptionRequest{
				CompanyID: companyID, OptionCode: gen.TripPublication,
			})
			require.Empty(t, publication.Errors)
			require.NotNil(t, publication.Data)
			require.Equal(t, tt.wantPublication, publication.Data.Allowed)
			if tt.wantPublication {
				require.Nil(t, publication.Data.Reason)
			} else {
				require.NotNil(t, publication.Data.Reason)
				require.Equal(t, gen.SERVICENOTALLOWED, *publication.Data.Reason)
			}
			creation := sendCheckOptionRequest(t, http.StatusOK, gen.CheckOptionRequest{
				CompanyID: companyID, OptionCode: gen.TripCreation,
			})
			require.Empty(t, creation.Errors)
			require.NotNil(t, creation.Data)
			require.Equal(t, tt.wantCreation, creation.Data.Allowed)
		})
	}
}
