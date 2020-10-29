package random

import (
	"context"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type server struct {
	providerSchema     *tfprotov5.Schema
	providerMetaSchema *tfprotov5.Schema
	resourceSchemas    map[string]*tfprotov5.Schema
	dataSourceSchemas  map[string]*tfprotov5.Schema

	tfprotov5.ResourceRouter
	tfprotov5.DataSourceRouter
}

func (s server) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	return &tfprotov5.GetProviderSchemaResponse{
		Provider:          s.providerSchema,
		ProviderMeta:      s.providerMetaSchema,
		ResourceSchemas:   s.resourceSchemas,
		DataSourceSchemas: s.dataSourceSchemas,
	}, nil
}

func (s server) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	return &tfprotov5.PrepareProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (s server) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	return &tfprotov5.ConfigureProviderResponse{}, nil
}

func (s server) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	return &tfprotov5.StopProviderResponse{}, nil
}

func Server() tfprotov5.ProviderServer {
	return server{
		providerSchema: &tfprotov5.Schema{
			Version: 1,
			Block:   &tfprotov5.SchemaBlock{},
		},
		dataSourceSchemas: map[string]*tfprotov5.Schema{},
		DataSourceRouter:  tfprotov5.DataSourceRouter{},
		resourceSchemas: map[string]*tfprotov5.Schema{
			"random_pet": {
				Version: 1,
				Block: &tfprotov5.SchemaBlock{
					Version: 1,
					Attributes: []*tfprotov5.SchemaAttribute{
						{
							Name:     "id",
							Type:     tftypes.String,
							Computed: true,
						},
						{
							Name:            "length",
							Type:            tftypes.Number,
							Description:     "The length (in words) of the pet name.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Computed:        true,
						},
						{
							Name:            "separator",
							Type:            tftypes.String,
							Description:     "The character to separate words in the pet name.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Computed:        true,
						},
						{
							Name: "component",
							NestedBlock: &tfprotov5.SchemaNestedBlock{
								TypeName: "component",
								Block: &tfprotov5.SchemaBlock{
									Attributes: []*tfprotov5.SchemaAttribute{
										{
											Type:        tftypes.String,
											Name:        "prefix",
											Optional:    true,
											Description: "a string prefix for the pet name",
										},
									},
								},
								Nesting: tfprotov5.SchemaNestedBlockNestingModeSingle,
							},
							Description:     "The character to separate words in the pet name.",
							DescriptionKind: tfprotov5.StringKindPlain,
							Optional:        true,
						},
					},
				},
			},
		},
		ResourceRouter: tfprotov5.ResourceRouter{
			"random_pet": resourcePet{},
		},
	}
}
