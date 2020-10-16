package random

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type resourcePet struct {
	Id        string    `json:"id"`
	Length    big.Float `json:"length"`
	Separator string    `json:"separator"`
	Component component `json:"component"`
}

func (r *resourcePet) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if id, ok := v["id"]; ok {
		r.Id = id.(string)
	}
	if len, ok := v["length"]; ok {
		float := big.NewFloat(len.(float64))
		r.Length = *float
	}
	if sep, ok := v["separator"]; ok {
		r.Separator = sep.(string)
	} else {
		r.Separator = "-"
	}
	if comp, ok := v["component"]; ok {
		jcomp, _ := json.Marshal(comp)
		var v map[string]interface{}
		json.Unmarshal(jcomp, &v)
		if prefix, ok := v["prefix"]; ok {
			r.Component.Prefix = prefix.(string)
		}
	}
	return nil
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
	err = val["separator"].As(&r.Separator)
	if err != nil {
		return err
	}
	if r.Separator == "" {
		r.Separator = "-"
	}
	err = val["component"].As(&r.Component)
	if err != nil {
		return err
	}
	return nil
}

type component struct {
	Prefix string `json:"prefix"`
}

func (c component) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"prefix": tftypes.String,
		},
	}
}

func (c *component) UnmarshalTerraform5Type(v tftypes.Value) error {
	var val map[string]tftypes.Value
	err := v.As(&val)
	if err != nil {
		return err
	}
	err = val["prefix"].As(&c.Prefix)
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
			"component": component{}.schema(),
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
	if r.Component.Prefix != "" {
		pet = fmt.Sprintf("%s%s%s", r.Component.Prefix, r.Separator, pet)
	}

	var comp tftypes.Value
	if r.Component.Prefix == "" {
		comp = tftypes.NewValue(component{}.schema(), nil)
	} else {
		comp = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
			"prefix": tftypes.NewValue(tftypes.String, r.Component.Prefix),
		})
	}

	state, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length":    tftypes.NewValue(tftypes.Number, &r.Length),
		"id":        tftypes.NewValue(tftypes.String, pet),
		"separator": tftypes.NewValue(tftypes.String, r.Separator),
		"component": comp,
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

	var comp tftypes.Value
	if r.Component.Prefix == "" {
		comp = tftypes.NewValue(component{}.schema(), nil)
	} else {
		comp = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
			"prefix": tftypes.NewValue(tftypes.String, r.Component.Prefix),
		})
	}

	proposedState, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, &r.Length),
		// Add an unknown value for id, so we can populate it during apply
		"id":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"component": comp,
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
	var state resourcePet
	err := json.Unmarshal(req.RawState.JSON, &state)
	if err != nil {
		panic(err)
	}

	var comp tftypes.Value
	if state.Component.Prefix == "" {
		comp = tftypes.NewValue(component{}.schema(), nil)
	} else {
		comp = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
			"prefix": tftypes.NewValue(tftypes.String, state.Component.Prefix),
		})
	}

	upgradedState, err := tftypes.NewValue(r.schema(), map[string]tftypes.Value{
		"length":    tftypes.NewValue(tftypes.Number, &state.Length),
		"id":        tftypes.NewValue(tftypes.String, state.Id),
		"separator": tftypes.NewValue(tftypes.String, state.Separator),
		"component": comp,
	}).MarshalMsgPack(r.schema())

	if err != nil {
		panic(err)
	}

	return &tfprotov5.UpgradeResourceStateResponse{
		UpgradedState: &tfprotov5.DynamicValue{
			MsgPack: upgradedState,
		},
	}, nil
}
