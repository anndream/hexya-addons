// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

import (
	"fmt"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/orm"
	"strings"
	"sync"
)

type ActionType string

const (
	ACTION_ACT_WINDOW ActionType = "ir.actions.act_window"
	ACTION_SERVER     ActionType = "ir.actions.server"
)

type ActionRef [2]string

func (e *ActionRef) String() string {
	sl := []string{e[0], e[1]}
	return fmt.Sprintf(`["%s"]`, strings.Join(sl, ","))
}

func (e *ActionRef) FieldType() int {
	return orm.TypeTextField
}

func (e *ActionRef) SetRaw(value interface{}) error {
	switch d := value.(type) {
	case string:
		dTrimmed := strings.Trim(d, "[]")
		tokens := strings.Split(dTrimmed, ",")
		if len(tokens) > 1 {
			*e = [2]string{tokens[0], tokens[1]}
			return nil
		}
		e = nil
		return fmt.Errorf("<ActionRef.SetRaw>Unable to parse %s", d)
	default:
		return fmt.Errorf("<ActionRef.SetRaw> unknown value `%v`", value)
	}
}

func (e *ActionRef) RawValue() interface{} {
	return e.String()
}

var _ orm.Fielder = new(ActionRef)

type ActionsCollection struct {
	sync.RWMutex
	actions map[string]*BaseAction
}

// NewActionCollection returns a pointer to a new
// ActionsCollection instance
func NewActionsCollection() *ActionsCollection {
	res := ActionsCollection{
		actions: make(map[string]*BaseAction),
	}
	return &res
}

// AddAction adds the given action to our ActionsCollection
func (ar *ActionsCollection) AddAction(a *BaseAction) {
	ar.Lock()
	defer ar.Unlock()
	ar.actions[a.ID] = a
}

// GetActionById returns the Action with the given id
func (ar *ActionsCollection) GetActionById(id string) *BaseAction {
	return ar.actions[id]
}

type BaseAction struct {
	ID     string     `json:"id"`
	Type   ActionType `json:"type"`
	Name   string     `json:"name"`
	Model  string     `json:"res_model"`
	ResID  int64      `json:"res_id"`
	Groups []string   `json:"groups_id"`
	Domain string     `json:"domain"`
	Help   string     `json:"help"`
	//SearchView *View `json:"search_view_id"`
	SrcModel string `json:"src_model"`
	Usage    string `json:"usage"`
	//Flags interface{}`json:"flags"`
	//Views []ViewRef `json:"views"`
	//View ViewRef `json:"view_id"`
	AutoRefresh bool           `json:"auto_refresh"`
	ViewMode    string         `json:"view_mode"`
	ViewIds     []string       `json:"view_ids"`
	Multi       bool           `json:"multi"`
	Target      string         `json:"target"`
	AutoSearch  bool           `json:"auto_search"`
	SearchView  string         `json:"search_view"`
	Filter      bool           `json:"filter"`
	Limit       int64          `json:"limit"`
	Context     models.Context `json:"context"`
}
