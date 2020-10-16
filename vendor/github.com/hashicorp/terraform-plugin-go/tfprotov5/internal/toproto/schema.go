package toproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func Schema(in *tfprotov5.Schema) (*tfplugin5.Schema, error) {
	var resp tfplugin5.Schema
	resp.Version = in.Version
	if in.Block != nil {
		block, err := Schema_Block(in.Block)
		if err != nil {
			return &resp, fmt.Errorf("error marshalling block: %w", err)
		}
		resp.Block = block
	}
	return &resp, nil
}

func Schema_Block(in *tfprotov5.SchemaBlock) (*tfplugin5.Schema_Block, error) {
	resp := &tfplugin5.Schema_Block{
		Version:         in.Version,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
		Sensitive:       in.Sensitive,
	}
	attrs, err := Schema_Attributes(in.Attributes)
	if err != nil {
		return resp, err
	}
	resp.Attributes = attrs
	blocks, err := Schema_NestedBlocks(in.BlockTypes)
	if err != nil {
		return resp, err
	}
	resp.BlockTypes = blocks
	return resp, nil
}

func Schema_Attribute(in *tfprotov5.SchemaAttribute) (*tfplugin5.Schema_Attribute, error) {
	resp := &tfplugin5.Schema_Attribute{
		Name:            in.Name,
		Description:     in.Description,
		Required:        in.Required,
		Optional:        in.Optional,
		Computed:        in.Computed,
		Sensitive:       in.Sensitive,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}
	t, err := CtyType(in.Type)
	if err != nil {
		return resp, fmt.Errorf("error marshaling type to JSON: %w", err)
	}
	resp.Type = t
	return resp, nil
}

func Schema_Attributes(in []*tfprotov5.SchemaAttribute) ([]*tfplugin5.Schema_Attribute, error) {
	resp := make([]*tfplugin5.Schema_Attribute, 0, len(in))
	for _, a := range in {
		if a == nil {
			resp = append(resp, nil)
			continue
		}
		attr, err := Schema_Attribute(a)
		if err != nil {
			return nil, err
		}
		resp = append(resp, attr)
	}
	return resp, nil
}

func Schema_NestedBlock(in *tfprotov5.SchemaNestedBlock) (*tfplugin5.Schema_NestedBlock, error) {
	resp := &tfplugin5.Schema_NestedBlock{
		TypeName: in.TypeName,
		Nesting:  Schema_NestedBlock_NestingMode(in.Nesting),
		MinItems: in.MinItems,
		MaxItems: in.MaxItems,
	}
	if in.Block != nil {
		block, err := Schema_Block(in.Block)
		if err != nil {
			return resp, fmt.Errorf("error marshaling nested block: %w", err)
		}
		resp.Block = block
	}
	return resp, nil
}

func Schema_NestedBlocks(in []*tfprotov5.SchemaNestedBlock) ([]*tfplugin5.Schema_NestedBlock, error) {
	resp := make([]*tfplugin5.Schema_NestedBlock, 0, len(in))
	for _, b := range in {
		if b == nil {
			resp = append(resp, nil)
			continue
		}
		block, err := Schema_NestedBlock(b)
		if err != nil {
			return nil, err
		}
		resp = append(resp, block)
	}
	return resp, nil
}

func Schema_NestedBlock_NestingMode(in tfprotov5.SchemaNestedBlockNestingMode) tfplugin5.Schema_NestedBlock_NestingMode {
	return tfplugin5.Schema_NestedBlock_NestingMode(in)
}
