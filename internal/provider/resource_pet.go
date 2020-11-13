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
	Id        string
	Length    *big.Float
	Separator string
	Prefix    string
}

func (r *resourcePet) FromTerraform5Value(v tftypes.Value) error {
	var val map[string]tftypes.Value
	err := v.As(&val)
	if err != nil {
		return err
	}

	err = val["id"].As(&r.Id)
	if err != nil {
		return err
	}

	r.Length = &big.Float{}
	err = val["length"].As(&r.Length)
	if err != nil {
		return err
	}
	if r.Length == nil {
		r.Length = big.NewFloat(float64(3))
	}

	err = val["separator"].As(&r.Separator)
	if err != nil {
		return err
	}
	if r.Separator == "" {
		r.Separator = "-"
	}

	err = val["prefix"].As(&r.Prefix)
	if err != nil {
		return err
	}

	return nil
}

func (r resourcePet) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length":    tftypes.Number,
			"id":        tftypes.String,
			"separator": tftypes.String,
			"prefix":    tftypes.String,
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
	if r.Prefix != "" {
		pet = fmt.Sprintf("%s%s%s", r.Prefix, r.Separator, pet)
	}

	state, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length":    tftypes.NewValue(tftypes.Number, r.Length),
		"id":        tftypes.NewValue(tftypes.String, pet),
		"separator": tftypes.NewValue(tftypes.String, r.Separator),
		"prefix":    tftypes.NewValue(tftypes.String, r.Prefix),
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

	// If we are doing a destroy, return the proposed state
	if state.IsNull() {
		return &tfprotov5.PlanResourceChangeResponse{
			PlannedState: req.ProposedNewState,
		}, nil
	}

	var id tftypes.Value
	if r.Id == "" {
		id = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	} else {
		id = tftypes.NewValue(tftypes.String, r.Id)
	}

	proposedState, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length":    tftypes.NewValue(tftypes.Number, r.Length),
		"id":        id,
		"prefix":    tftypes.NewValue(tftypes.String, r.Prefix),
		"separator": tftypes.NewValue(tftypes.String, r.Separator),
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
	return &tfprotov5.ReadResourceResponse{
		NewState: &tfprotov5.DynamicValue{
			MsgPack: req.CurrentState.MsgPack,
		},
	}, nil
}

func (r resourcePet) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	return &tfprotov5.UpgradeResourceStateResponse{
		UpgradedState: &tfprotov5.DynamicValue{
			JSON: req.RawState.JSON,
		},
	}, nil
}
