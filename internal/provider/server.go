package random

import (
	"context"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type server struct {
	providerSchema     *tfprotov6.Schema
	providerMetaSchema *tfprotov6.Schema
	resourceSchemas    map[string]*tfprotov6.Schema
	dataSourceSchemas  map[string]*tfprotov6.Schema

	tfprotov6.DataSourceServer
	tfprotov6.ResourceServer
}

func (s server) GetProviderSchema(ctx context.Context, req *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	return &tfprotov6.GetProviderSchemaResponse{
		Provider:          s.providerSchema,
		ProviderMeta:      s.providerMetaSchema,
		ResourceSchemas:   s.resourceSchemas,
		DataSourceSchemas: s.dataSourceSchemas,
	}, nil
}

func (s server) ValidateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	return &tfprotov6.ValidateProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (s server) ConfigureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	return &tfprotov6.ConfigureProviderResponse{}, nil
}

func (s server) StopProvider(ctx context.Context, req *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	return &tfprotov6.StopProviderResponse{}, nil
}

func Server() tfprotov6.ProviderServer {
	return server{
		providerSchema: &tfprotov6.Schema{
			Version: 1,
			Block:   &tfprotov6.SchemaBlock{},
		},
		resourceSchemas: map[string]*tfprotov6.Schema{
			"random_pet": {
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Version: 1,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "id",
							Type:     tftypes.String,
							Computed: true,
						},
						{
							Name:            "length",
							Type:            tftypes.Number,
							Description:     "The length (in words) of the pet name.",
							DescriptionKind: tfprotov6.StringKindPlain,
							Optional:        true,
							Computed:        true,
						},
						{
							Name:            "separator",
							Type:            tftypes.String,
							Description:     "The character to separate words in the pet name.",
							DescriptionKind: tfprotov6.StringKindPlain,
							Computed:        true,
						},
						{
							Name: "components",
							NestedType: &tfprotov6.SchemaObject{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Type:        tftypes.String,
										Name:        "prefix",
										Optional:    true,
										Description: "a string prefix for the pet name",
										Sensitive:   false,
									},
									{
										Type:        tftypes.String,
										Name:        "secret",
										Optional:    true,
										Description: "a string prefix for the pet name",
										Sensitive:   true,
									},
									{
										Type:        tftypes.String,
										Name:        "computed",
										Optional:    true,
										Description: "a string prefix for the pet name",
										Computed:    true,
									},
								},

								Nesting: tfprotov6.SchemaObjectNestingModeList,
							},
							Description:     "The character to separate words in the pet name.",
							DescriptionKind: tfprotov6.StringKindPlain,
							Optional:        true,
						},
					},
				},
			},
		},
		ResourceServer: resourcePet{},
	}
}
