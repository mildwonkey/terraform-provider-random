package random

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type resourcePet struct {
	Length big.Float
	Id     string
}

func (r *resourcePet) UnmarshalTerraform5Type(v tftypes.Value) error {
	var val map[string]tftypes.Value
	err := v.As(&val)
	if err != nil {
		return err
	}
	err = val["length"].As(&r.Length)
	if err != nil {
		return err
	}
	return nil
}

func (r resourcePet) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length": tftypes.Number,
			"id":     tftypes.String,
			//"separator": tftypes.String,
			// "component": tftypes.Object{
			// 	AttributeTypes: map[string]tftypes.Type{
			// 		"prefix": tftypes.String,
			// 	},
			// },
		},
	}
}

var (
	_ tfprotov5.ResourceServer = (*resourcePet)(nil)
)

func (r resourcePet) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	return &tfprotov5.ValidateResourceTypeConfigResponse{}, nil
}

func (r resourcePet) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	schema := r.schema()

	config, err := req.Config.Unmarshal(schema)
	if err != nil {
		panic(err)
	}
	err = config.As(&r)
	if err != nil {
		panic(err)
	}

	length, _ := r.Length.Int64()
	pet := strings.ToLower(petname.Generate(int(length), "-"))

	state, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, &r.Length),
		"id":     tftypes.NewValue(tftypes.String, pet),
	}).MarshalMsgPack(schema)

	if err != nil {
		return &tfprotov5.ApplyResourceChangeResponse{
			Diagnostics: []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "Error encoding state",
					Detail:   fmt.Sprintf("Error encoding state: %s", err.Error()),
				},
			},
		}, nil
	}

	return &tfprotov5.ApplyResourceChangeResponse{
		NewState: &tfprotov5.DynamicValue{
			MsgPack: state,
		},
	}, nil
}

func (r resourcePet) ImportResourceState(ctx context.Context, req *tfprotov5.ImportResourceStateRequest) (*tfprotov5.ImportResourceStateResponse, error) {
	return &tfprotov5.ImportResourceStateResponse{}, nil
}

func (r resourcePet) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	schema := r.schema()

	state, err := req.ProposedNewState.Unmarshal(schema)
	if err != nil {
		panic(err)
	}
	err = state.As(&r)
	if err != nil {
		panic(err)
	}

	// Add an unknown value for id, so we can populate it during apply
	proposedState, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, &r.Length),
		"id":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	}).MarshalMsgPack(schema)
	if err != nil {
		panic(err)
	}

	return &tfprotov5.PlanResourceChangeResponse{
		PlannedState: &tfprotov5.DynamicValue{
			MsgPack: proposedState,
		},
	}, nil
}

func (r resourcePet) ReadResource(ctx context.Context, req *tfprotov5.ReadResourceRequest) (*tfprotov5.ReadResourceResponse, error) {
	return &tfprotov5.ReadResourceResponse{}, nil
}

func (r resourcePet) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	return &tfprotov5.UpgradeResourceStateResponse{
		UpgradedState: req.RawState,
	}, nil
}
