package optional_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hexastack-dev/devkit-go/data/optional"
	"github.com/stretchr/testify/assert"
)

type preference struct {
	Push             optional.Value[bool] `json:"push"`
	TransactionLimit optional.Value[int]  `json:"transactionLimit"`
}

type person struct {
	FirstName     optional.Value[string]      `json:"firstName,omitempty"`
	MiddleName    optional.Value[string]      `json:"middleName,omitempty"`
	LastName      optional.Value[string]      `json:"lastName,omitempty"`
	PreferredName optional.Value[string]      `json:"preferredName,omitempty"`
	Preference1   optional.Value[preference]  `json:"preference1,omitempty"`
	Preference2   optional.Value[preference]  `json:"preference2,omitempty"`
	Preference3   optional.Value[preference]  `json:"preference3,omitempty"`
	Preference4   optional.Value[preference]  `json:"preference4,omitempty"`
	Preference5   optional.Value[*preference] `json:"preference5,omitempty"`
	Preference6   optional.Value[*preference] `json:"preference6,omitempty"`
}

func TestNil(t *testing.T) {
	v1 := optional.Nil[int]()
	if !v1.IsNil() {
		t.Error("should be nil")
	}
	if !v1.IsDefined() {
		t.Error("should be defined")
	}
	if v1.Val() != 0 {
		t.Errorf("should be equals 0: %d", v1.Val())
	}
	if v1.String() != "nil" {
		t.Errorf("String() should be equals nil: %s", v1.String())
	}

	v2 := optional.Nil[string]()
	if !v2.IsNil() {
		t.Error("should be nil")
	}
	if !v2.IsDefined() {
		t.Error("should be defined")
	}
	if v2.Val() != "" {
		t.Errorf("Value() should be empty string: %s", v2.Val())
	}
	if v2.String() != "nil" {
		t.Errorf("String() should be equals nil: %s", v2.String())
	}

	v3 := optional.Nil[person]()
	if !v3.IsNil() {
		t.Error("should be nil")
	}
	if !v3.IsDefined() {
		t.Error("should be defined")
	}
	if v3.Val() != (person{}) {
		t.Errorf("Value() should be empty struct: %+v", v3.Val())
	}
	if v3.String() != "nil" {
		t.Errorf("String() should be equals nil: %s", v3.String())
	}

	v4 := optional.Nil[*person]()
	if !v4.IsNil() {
		t.Error("should be nil")
	}
	if !v4.IsDefined() {
		t.Error("should be defined")
	}
	if v4.Val() != nil {
		t.Errorf("Value() should be nil: %+v", v4.Val())
	}
	if v4.String() != "nil" {
		t.Errorf("String() should be equals nil: %s", v4.String())
	}
}

func TestOf(t *testing.T) {
	v1 := optional.Of(32)
	if v1.IsNil() {
		t.Errorf("should not be nil")
	}
	if !v1.IsDefined() {
		t.Error("should be defined")
	}
	if v1.Val() != 32 {
		t.Errorf("Value() should be equals 32: %d", v1.Val())
	}
	if v1.String() != "32" {
		t.Errorf(`String() should be equals "32": %s`, v1.String())
	}

	v2 := optional.Of("hello")
	if v2.IsNil() {
		t.Errorf("should not be nil")
	}
	if !v2.IsDefined() {
		t.Error("should be defined")
	}
	if v2.Val() != "hello" {
		t.Errorf("Value() should be equals hello: %s", v2.Val())
	}
	if v2.String() != "hello" {
		t.Errorf(`String() should be equals hello: %s`, v2.String())
	}

	v3 := optional.Of(person{})
	if v3.IsNil() {
		t.Errorf("should not be nil")
	}
	if !v3.IsDefined() {
		t.Error("should be defined")
	}
	if v3.Val() != (person{}) {
		t.Errorf("Value() should be equals empty struct: %+v", v3.Val())
	}
	if v3.String() == "{}" {
		t.Errorf("String() should not be equals empty string")
	}

	v4 := optional.Of[*person](nil)
	if v4.IsNil() {
		t.Errorf("v4.IsNil() should return false: %t", v4.IsNil())
	} else {
		fmt.Printf("v4.IsNil() return %t This is wrong but expected to avoid usage of reflection, see the doc for more info.\n", v4.IsNil())
	}
}

var jsonBlob = []byte(`{
	"firstName": "Tommy",
	"middleName": "",
	"lastName": null,
	"preference1": {
		"push": false,
		"transactionLimit": 0
	},
	"preference2": {
		"push": true,
		"transactionLimit": 1000
	},
	"preference3": {},
	"preference5": {
		"transactionLimit": 5000
	}
}`)

func (p *person) MarshalJSON() ([]byte, error) {
	p2 := struct {
		FirstName     *string        `json:"firstName,omitempty"`
		MiddleName    *string        `json:"middleName,omitempty"`
		LastName      *string        `json:"lastName,omitempty"`
		PreferredName *string        `json:"preferredName,omitempty"`
		Preference1   *preference    `json:"preference1,omitempty"`
		Preference2   *preference    `json:"preference2,omitempty"`
		Preference3   *preference    `json:"preference3,omitempty"`
		Preference4   *preference    `json:"preference4,omitempty"`
		Preference5   *(*preference) `json:"preference5,omitempty"`
		Preference6   *(*preference) `json:"preference6,omitempty"`
	}{
		FirstName:     p.FirstName.ValuePtr(),
		MiddleName:    p.MiddleName.ValuePtr(),
		LastName:      p.LastName.ValuePtr(),
		PreferredName: p.PreferredName.ValuePtr(),
		Preference1:   p.Preference1.ValuePtr(),
		Preference2:   p.Preference2.ValuePtr(),
		Preference3:   p.Preference3.ValuePtr(),
		Preference4:   p.Preference4.ValuePtr(),
		Preference5:   p.Preference5.ValuePtr(),
		Preference6:   p.Preference6.ValuePtr(),
	}

	return json.Marshal(p2)
}

func TestOptional_UnmarshalJSON(t *testing.T) {
	var p person
	if err := json.Unmarshal(jsonBlob, &p); err != nil {
		t.Fatal(err)
	}

	if p.FirstName.IsNil() {
		t.Errorf("FirstName should not be nil")
	}
	if !p.FirstName.IsDefined() {
		t.Errorf("FirstName should be defined")
	}

	if p.MiddleName.IsNil() {
		t.Errorf("MiddleName should not be nil")
	}
	if !p.MiddleName.IsDefined() {
		t.Errorf("FirstName should be defined")
	}

	if !p.LastName.IsNil() {
		t.Errorf("LastName should be nil")
	}
	if !p.LastName.IsDefined() {
		t.Errorf("LastName should be defined")
	}

	if !p.PreferredName.IsNil() {
		t.Errorf("PreferredName should be nil")
	}
	if p.PreferredName.IsDefined() {
		t.Errorf("PreferredName should be undefined")
	}

	if p.Preference1.IsNil() {
		t.Errorf("Preference1 should not be nil")
	} else {
		pref := p.Preference1.Val()
		if pref.Push.IsNil() {
			t.Errorf("Preference1.Push should not be nil")
		} else {
			fmt.Printf("Preference1.Push: \"%t\"\n", pref.Push.Val())
		}

		if pref.TransactionLimit.IsNil() {
			t.Errorf("Preference1.TransactionLimit should not be nil")
		} else {
			fmt.Printf("Preference1.TransactionLimit: \"%d\"\n", pref.TransactionLimit.Val())
		}
	}

	if p.Preference2.IsNil() {
		t.Errorf("Preference2 should not be nil")
	} else {
		pref := p.Preference2.Val()
		if pref.Push.IsNil() {
			t.Errorf("Preference2.Push should not be nil")
		} else if !pref.Push.Val() {
			t.Errorf("Preference2.Push should be true")
		} else {
			fmt.Printf("Preference2.Push: \"%t\"\n", pref.Push.Val())
		}

		if pref.TransactionLimit.IsNil() {
			t.Errorf("Preference2.TransactionLimit should not be nil")
		} else if pref.TransactionLimit.Val() == 0 {
			t.Errorf("Preference2.TransactionLimit should not equals 0")
		} else {
			fmt.Printf("Preference2.TransactionLimit: \"%d\"\n", pref.TransactionLimit.Val())
		}
	}

	if p.Preference3.IsNil() {
		t.Errorf("Preference3 should not be nil")
	} else {
		pref := p.Preference3.Val()
		if !pref.Push.IsNil() {
			t.Errorf("Preference2.Push should be nil")
		}
		if !pref.TransactionLimit.IsNil() {
			t.Errorf("Preference2.TransactionLimit should be nil")
		}
	}

	if !p.Preference4.IsNil() {
		t.Errorf("Preference4 should be nil")
	}

	if p.Preference5.IsNil() {
		t.Errorf("Preference5 should not be nil")
	} else {
		pref := p.Preference5.Val()
		if pref.TransactionLimit.IsNil() {
			t.Errorf("Preference5.TransactionLimit should not be nil")
		} else {
			fmt.Printf("Preference5.TransactionLimit: \"%d\"\n", pref.TransactionLimit.Val())
		}
	}

	if !p.Preference6.IsNil() {
		t.Errorf("Preference6 should be nil")
	}
}

/**
{
	"firstName": "Tommy",
	"middleName": "",
	"lastName": null,
	"preference1": {
		"push": false,
		"transactionLimit": 0
	},
	"preference2": {
		"push": true,
		"transactionLimit": 1000
	},
	"preference3": {},
	"preference5": {
		"transactionLimit": 5000
	}
}
*/

func TestOptional_MarshalJSON(t *testing.T) {
	var p person
	p.FirstName = optional.Of("Tommy")
	p.MiddleName = optional.Of("")
	p.LastName = optional.Nil[string]()
	p.Preference1 = optional.Of(preference{
		Push:             optional.Of(false),
		TransactionLimit: optional.Of(0),
	})
	p.Preference2 = optional.Of(preference{
		Push:             optional.Of(true),
		TransactionLimit: optional.Of(1000),
	})
	p.Preference3 = optional.Of(preference{})
	p.Preference5 = optional.Of(&preference{
		TransactionLimit: optional.Of(5000),
	})

	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	bs := string(b)
	if bs != `{"firstName":"Tommy","middleName":"","preference1":{"push":false,"transactionLimit":0},"preference2":{"push":true,"transactionLimit":1000},"preference3":{"push":false,"transactionLimit":0},"preference5":{"push":false,"transactionLimit":5000}}` {
		t.Errorf("marshal result should be equals, got: %s", bs)
	}
	fmt.Printf("%s\n", b)
}

func TestScan(t *testing.T) {
	var (
		test1 optional.Value[string]
		test2 optional.Value[int]
		test3 optional.Value[string]
		test4 optional.Value[string]

		val1     = "hello"
		val2     = 32
		val3 any = nil
	)

	assert.NoError(t, test1.Scan(val1))
	assert.False(t, test1.IsNil())
	assert.True(t, test1.IsDefined())
	assert.Equal(t, "hello", test1.Val())

	assert.NoError(t, test2.Scan(val2))
	assert.False(t, test2.IsNil())
	assert.True(t, test2.IsDefined())
	assert.Equal(t, 32, test2.Val())

	assert.NoError(t, test3.Scan(val3))
	assert.True(t, test3.IsNil())
	assert.True(t, test3.IsDefined())
	assert.Zero(t, test3.Val())

	err := test4.Scan(val2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, optional.ErrTypeMismatch)
	assert.True(t, test4.IsNil())
	assert.True(t, test4.IsDefined())
	assert.Zero(t, test4.Val())
}
