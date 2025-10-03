/*
Copyright 2024-present jishaocong0910

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gdao_test

import (
	"testing"

	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

func TestSubPrefix(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().SubPrefix("Tb")
	r.Equal("", mapper.Convert(""))
	r.Equal("Product", mapper.Convert("TbProduct"))
}

func TestSubSuffix(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().SubSuffix("Po")
	r.Equal("", mapper.Convert(""))
	r.Equal("Product", mapper.Convert("ProductPo"))
}

func TestAddPrefix(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().AddPrefix("Tb")
	r.Equal("", mapper.Convert(""))
	r.Equal("TbProduct", mapper.Convert("Product"))
}

func TestAddSuffix(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().AddSuffix("Po")
	r.Equal("", mapper.Convert(""))
	r.Equal("ProductPo", mapper.Convert("Product"))
}

func TestLowerCamelCase(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().LowerCamelCase()
	r.Equal("", mapper.Convert(""))
	r.Equal("tbProductPo", mapper.Convert("tb_product_Po"))
	r.Equal("tbProductPo", mapper.Convert("Tb Product Po"))
	r.Equal("tbProductPo", mapper.Convert("Tb-Product-Po"))
	r.Equal("tbProductPo", mapper.Convert("TB_PRODUCT_PO"))
	r.Equal("tbProductPo", mapper.Convert("TbProductPo"))
	r.Equal("tbProductPo", mapper.Convert("tbProductPo"))
}

func TestLowerSnakeCase(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().LowerSnakeCase()
	r.Equal("", mapper.Convert(""))
	r.Equal("tb_product_po", mapper.Convert("TbProductPo"))
	r.Equal("tb_product_po", mapper.Convert("Tb Product Po"))
	r.Equal("tb_product_po", mapper.Convert("Tb-Product-Po"))
	r.Equal("tb_product_po", mapper.Convert("TB_PRODUCT_PO"))
}

func TestLowerFirstLiteral(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().LowerFirstLiteral()
	r.Equal("", mapper.Convert(""))
	r.Equal("tbProductPo", mapper.Convert("TbProductPo"))
}

func TestUpperCamelCase(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().UpperCamelCase()
	r.Equal("", mapper.Convert(""))
	r.Equal("TbProductPo", mapper.Convert("tb_product_Po"))
	r.Equal("TbProductPo", mapper.Convert("Tb Product Po"))
	r.Equal("TbProductPo", mapper.Convert("Tb-Product-Po"))
	r.Equal("TbProductPo", mapper.Convert("TB_PRODUCT_PO"))
	r.Equal("TbProductPo", mapper.Convert("TbProductPo"))
	r.Equal("TbProductPo", mapper.Convert("tbProductPo"))
}

func TestUpperSnakeCase(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().UpperSnakeCase()
	r.Equal("", mapper.Convert(""))
	r.Equal("TB_PRODUCT_PO", mapper.Convert("TbProductPo"))
	r.Equal("TB_PRODUCT_PO", mapper.Convert("Tb Product Po"))
	r.Equal("TB_PRODUCT_PO", mapper.Convert("Tb-Product-Po"))
	r.Equal("TB_PRODUCT_PO", mapper.Convert("tb_product_po"))
}

func TestUpperFirstLiteral(t *testing.T) {
	r := require.New(t)
	mapper := gdao.NewNameMapper().UpperFirstLiteral()
	r.Equal("", mapper.Convert(""))
	r.Equal("TbProductPo", mapper.Convert("tbProductPo"))
}
