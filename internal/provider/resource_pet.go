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
	Length    big.Float
	Separator string
	Component *component
}

func (r *resourcePet) FromTerraform5Value(v tftypes.Value) error {
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
	if comp, ok := val["component"]; ok {
		r.Component = &component{}
		err = comp.As(r.Component)
		if err != nil {
			return err
		}
	}
	return nil
}

type component struct {
	Prefix string
	List   []string
	Secret string
}

func (c component) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"prefix": tftypes.String,
			"list":   tftypes.List{ElementType: tftypes.String},
			"secret": tftypes.String,
		},
		OptionalAttrs: []string{"list", "secret"},
	}
}

func (c *component) FromTerraform5Value(v tftypes.Value) error {
	var val map[string]tftypes.Value
	err := v.As(&val)
	if err != nil {
		return err
	}
	// Prefix is required
	err = val["prefix"].As(&c.Prefix)
	if err != nil {
		return err
	}
	// optional, computed
	if listVal, ok := val["list"]; ok {
		var list []tftypes.Value
		err = listVal.As(&list)
		if err != nil {
			return err
		}
		stringlist := make([]string, len(list))
		for i, v := range list {
			var str string
			err = v.As(&str)
			stringlist[i] = str
		}
		c.List = stringlist
	}
	// optional
	if secretVal, ok := val["secret"]; ok {
		var secret string
		err = secretVal.As(&secret)
		if err != nil {
			return err
		}
		c.Secret = secret
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
	planned, err := req.PlannedState.Unmarshal(schema)
	// destroy
	if planned.IsNull() {
		return &tfprotov5.ApplyResourceChangeResponse{
			NewState: &tfprotov5.DynamicValue{
				MsgPack: req.PlannedState.MsgPack,
			},
		}, nil
	}

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
	if r.Component == nil {
		comp = tftypes.NewValue(component{}.schema(), nil)
	} else {
		var list, prefix, secret tftypes.Value
		prefix = tftypes.NewValue(tftypes.String, r.Component.Prefix)
		if r.Component.List != nil {
			listVals := make([]tftypes.Value, len(r.Component.List))
			for i, v := range r.Component.List {
				listVals[i] = tftypes.NewValue(tftypes.String, v)
			}
			list = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, listVals)
		} else {
			listVals := strings.Split(r.Separator, pet)
			list = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, listVals)
		}
		if r.Component.Secret != "" {
			secret = tftypes.NewValue(tftypes.String, r.Component.Secret)
		}

		comp = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
			"prefix": prefix,
			"list":   list,
			"secret": secret,
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
	// If we are doing a destroy, return the proposed state
	if state.IsNull() {
		return &tfprotov5.PlanResourceChangeResponse{
			PlannedState: req.ProposedNewState,
		}, nil
	}

	err = state.As(&r)
	if err != nil {
		panic(err)
	}

	var comp tftypes.Value
	if r.Component == nil {
		comp = tftypes.NewValue(component{}.schema(), nil)
	} else {
		var list, prefix, secret tftypes.Value
		prefix = tftypes.NewValue(tftypes.String, r.Component.Prefix)
		if r.Component.List != nil {
			listVals := make([]tftypes.Value, len(r.Component.List))
			for i, v := range r.Component.List {
				listVals[i] = tftypes.NewValue(tftypes.String, v)
			}
			list = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, listVals)
		} else {
			list = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue)
		}
		if r.Component.Secret != "" {
			secret = tftypes.NewValue(tftypes.String, r.Component.Secret)
		}

		comp = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
			"prefix": prefix,
			"list":   list,
			"secret": secret,
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
	return &tfprotov5.UpgradeResourceStateResponse{
		UpgradedState: &tfprotov5.DynamicValue{
			JSON: req.RawState.JSON,
		},
	}, nil
}
