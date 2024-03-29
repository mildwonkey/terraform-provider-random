package random

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type resourcePet struct {
	Id         string
	Length     int
	Separator  string
	Components components
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

	bigLen := big.Float{}
	err = val["length"].As(&bigLen)
	if err != nil {
		return err
	}
	length, _ := bigLen.Int64()
	if length == 0 {
		r.Length = 3
	} else {
		r.Length = int(length)
	}

	err = val["separator"].As(&r.Separator)
	if err != nil {
		return err
	}
	if r.Separator == "" {
		r.Separator = "-"
	}

	err = val["components"].As(&r.Components)
	if err != nil {
		return err
	}
	return nil
}

type components []component

func (c components) schema() tftypes.List {
	return tftypes.List{
		ElementType: component{}.schema(),
	}
}

func (c *components) FromTerraform5Value(v tftypes.Value) error {
	var values []tftypes.Value
	err := v.As(&values)
	if err != nil {
		return err
	}
	results := make([]component, 0, len(values))
	for _, val := range values {
		result := component{}
		err = val.As(&result)
		if err != nil {
			return err
		}
		results = append(results, result)
	}
	*c = results
	return nil
}

type component struct {
	Prefix   string
	Secret   string
	Computed string
}

func (c *component) FromTerraform5Value(v tftypes.Value) error {
	var val map[string]tftypes.Value
	err := v.As(&val)
	if err != nil {
		return err
	}
	err = val["prefix"].As(&c.Prefix)
	if err != nil {
		return err
	}
	if _, ok := val["secret"]; ok {
		err = val["secret"].As(&c.Secret)
		if err != nil {
			return err
		}
	}
	if _, ok := val["computed"]; ok {
		err = val["computed"].As(&c.Computed)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c component) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"prefix":   tftypes.String,
			"secret":   tftypes.String,
			"computed": tftypes.String,
		},
	}
}

func (r resourcePet) schema() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length":     tftypes.Number,
			"id":         tftypes.String,
			"separator":  tftypes.String,
			"components": components{}.schema(),
		},
	}
}

var (
	_ tfprotov6.ResourceServer = (*resourcePet)(nil)
)

func (r resourcePet) ValidateResourceConfig(ctx context.Context, req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	return &tfprotov6.ValidateResourceConfigResponse{}, nil
}

func (r resourcePet) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	schema := r.schema()
	planned, err := req.PlannedState.Unmarshal(schema)
	// destroy
	if planned.IsNull() {
		return &tfprotov6.ApplyResourceChangeResponse{
			NewState: &tfprotov6.DynamicValue{
				MsgPack: req.PlannedState.MsgPack,
			},
		}, nil
	}
	var planState resourcePet
	err = planned.As(&planState)

	config, err := req.Config.Unmarshal(schema)
	if err != nil {
		panic(err)
	}
	err = config.As(&r)
	if err != nil {
		panic(err)
	}

	var pet string
	if r.Id != "" {
		pet = r.Id
	} else {
		pet = strings.ToLower(petname.Generate(r.Length, "-"))
	}

	if len(r.Components) != 0 {
		for i, comp := range r.Components {
			pet = fmt.Sprintf("%s%s%s", comp.Prefix, r.Separator, pet)
			if comp.Computed == "" {
				r.Components[i].Computed = pet
			}
		}
	}

	state, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length":     tftypes.NewValue(tftypes.Number, &r.Length),
		"id":         tftypes.NewValue(tftypes.String, pet),
		"separator":  tftypes.NewValue(tftypes.String, r.Separator),
		"components": r.Components.ToTfValue(),
	}).MarshalMsgPack(schema)

	if err != nil {
		return &tfprotov6.ApplyResourceChangeResponse{
			Diagnostics: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "Error encoding state",
					Detail:   fmt.Sprintf("Error encoding state: %s", err.Error()),
				},
			},
		}, nil
	}

	return &tfprotov6.ApplyResourceChangeResponse{
		NewState: &tfprotov6.DynamicValue{
			MsgPack: state,
		},
	}, nil
}

func (r resourcePet) ImportResourceState(ctx context.Context, req *tfprotov6.ImportResourceStateRequest) (*tfprotov6.ImportResourceStateResponse, error) {
	return &tfprotov6.ImportResourceStateResponse{}, nil
}

func (r resourcePet) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	schema := r.schema()

	state, err := req.ProposedNewState.Unmarshal(schema)
	if err != nil {
		panic(err)
	}
	err = state.As(&r)
	if err != nil {
		panic(err)
	}

	// If we are doing a destroy, return the proposed state as-is
	if state.IsNull() {
		return &tfprotov6.PlanResourceChangeResponse{
			PlannedState: req.ProposedNewState,
		}, nil
	}

	var id tftypes.Value
	if r.Id == "" {
		id = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	} else {
		prior, err := req.PriorState.Unmarshal(schema)
		if err != nil {
			panic(err)
		}
		var priorPet resourcePet
		err = prior.As(&priorPet)
		if err != nil {
			panic(err)
		}
		if r.Equals(priorPet) {
			id = tftypes.NewValue(tftypes.String, r.Id)
		} else {
			id = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		}
	}

	proposedState, err := tftypes.NewValue(schema, map[string]tftypes.Value{
		"length":     tftypes.NewValue(tftypes.Number, &r.Length),
		"id":         id,
		"components": r.Components.ToTfValue(),
		"separator":  tftypes.NewValue(tftypes.String, r.Separator),
	}).MarshalMsgPack(schema)
	if err != nil {
		return &tfprotov6.PlanResourceChangeResponse{
			Diagnostics: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "Error encoding state",
					Detail:   fmt.Sprintf("Error encoding state: %s", err.Error()),
				},
			},
		}, nil
	}

	return &tfprotov6.PlanResourceChangeResponse{
		PlannedState: &tfprotov6.DynamicValue{
			MsgPack: proposedState,
		},
	}, nil
}

func (r resourcePet) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	return &tfprotov6.ReadResourceResponse{
		NewState: &tfprotov6.DynamicValue{
			MsgPack: req.CurrentState.MsgPack,
		},
	}, nil
}

func (r resourcePet) UpgradeResourceState(ctx context.Context, req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	return &tfprotov6.UpgradeResourceStateResponse{
		UpgradedState: &tfprotov6.DynamicValue{
			JSON: req.RawState.JSON,
		},
	}, nil
}

func (c components) ToTfValue() tftypes.Value {
	var comp tftypes.Value
	if len(c) == 0 {
		comp = tftypes.NewValue(components{}.schema(), nil)
	} else {
		list := make([]tftypes.Value, len(c))
		for i, v := range c {
			// optional values
			var secret, computed tftypes.Value
			if v.Secret != "" {
				secret = tftypes.NewValue(tftypes.String, v.Secret)
			} else {
				secret = tftypes.NewValue(tftypes.String, nil)
			}
			if v.Computed != "" {
				computed = tftypes.NewValue(tftypes.String, v.Computed)
			} else {
				computed = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			}

			list[i] = tftypes.NewValue(component{}.schema(), map[string]tftypes.Value{
				"prefix":   tftypes.NewValue(tftypes.String, v.Prefix),
				"secret":   secret,
				"computed": computed,
			})
		}
		comp = tftypes.NewValue(components{}.schema(), list)
	}
	return comp
}

func (r resourcePet) Equals(other resourcePet) bool {
	// TODO: write an actual comparison, this is just for hackery
	return reflect.DeepEqual(r, other)
}
