// Copyright 2012 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"

	"github.com/globalsign/mgo/bson"
	authTypes "github.com/tsuru/tsuru/types/auth"
	check "gopkg.in/check.v1"
)

func (s *S) TestTeamServiceCreate(c *check.C) {
	teamName := "pos"
	tags := []string{"tag1", "tag1 ", "tag2"}
	one := authTypes.User{Email: "king@pos.com"}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnInsert: func(t authTypes.Team) error {
				c.Assert(t.Name, check.Equals, teamName)
				c.Assert(t.CreatingUser, check.DeepEquals, one.Email)
				c.Assert(t.Tags, check.DeepEquals, []string{"tag1", "tag2"})
				return nil
			},
		},
	}

	err := ts.Create(context.TODO(), teamName, tags, &one)
	c.Assert(err, check.IsNil)
}

func (s *S) TestTeamServiceUpdate(c *check.C) {
	teamName := "pos"
	tags := []string{"tag1", "tag1 ", "tag2"}
	one := authTypes.User{Email: "king@pos.com"}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnFindByName: func(name string) (*authTypes.Team, error) {
				return &authTypes.Team{Name: teamName, Tags: []string{"tag3"}, CreatingUser: one.Email}, nil
			},
			OnUpdate: func(t authTypes.Team) error {
				c.Assert(t.Name, check.Equals, teamName)
				c.Assert(t.CreatingUser, check.DeepEquals, one.Email)
				c.Assert(t.Tags, check.DeepEquals, []string{"tag1", "tag2"})
				return nil
			},
		},
	}
	err := ts.Update(context.TODO(), teamName, tags)
	c.Assert(err, check.IsNil)
}

func (s *S) TestTeamServiceCreateDuplicate(c *check.C) {
	teamName := "pos"
	u := authTypes.User{Email: "king@pos.com"}
	tags := []string{"tag1=val1"}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnInsert: func(t authTypes.Team) error {
				c.Assert(t.Name, check.Equals, teamName)
				c.Assert(t.CreatingUser, check.DeepEquals, u.Email)
				c.Assert(t.Tags, check.DeepEquals, tags)
				return authTypes.ErrTeamAlreadyExists
			},
		},
	}
	err := ts.Create(context.TODO(), "pos", tags, &u)
	c.Assert(err, check.Equals, authTypes.ErrTeamAlreadyExists)
}

func (s *S) TestTeamServiceCreateTrimsName(c *check.C) {
	u := authTypes.User{Email: "king@pos.com"}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnInsert: func(t authTypes.Team) error {
				c.Assert(t.Name, check.Equals, "pos")
				return nil
			},
		},
	}

	err := ts.Create(context.TODO(), "pos", nil, &u)
	c.Assert(err, check.IsNil)
}

func (s *S) TestTeamServiceCreateValidation(c *check.C) {
	u := authTypes.User{Email: "king@pos.com"}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnInsert: func(t authTypes.Team) error {
				return nil
			},
		},
	}
	var tests = []struct {
		input string
		err   error
	}{
		{"", authTypes.ErrInvalidTeamName},
		{"    ", authTypes.ErrInvalidTeamName},
		{"1abc", authTypes.ErrInvalidTeamName},
		{"@abc", authTypes.ErrInvalidTeamName},
		{"my team", authTypes.ErrInvalidTeamName},
		{"Abacaxi", authTypes.ErrInvalidTeamName},
		{"TEAM", authTypes.ErrInvalidTeamName},
		{"TeaM", authTypes.ErrInvalidTeamName},
		{"team_1", nil},
		{"tsuru@corp.globo.com", nil},
		{"team-1", nil},
		{"a", authTypes.ErrInvalidTeamName},
		{"ab", nil},
		{"team1", nil},
	}

	for _, t := range tests {
		err := ts.Create(context.TODO(), t.input, nil, &u)
		if err != t.err {
			c.Errorf("Is %q valid? Want %v. Got %v.", t.input, t.err, err)
		}
	}
}

func (s *S) TestTeamServiceRemove(c *check.C) {
	teamName := "atreides"
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnDelete: func(t authTypes.Team) error {
				c.Assert(t.Name, check.Equals, teamName)
				return nil
			},
		},
	}

	err := ts.Remove(context.TODO(), teamName)
	c.Assert(err, check.IsNil)
}

func (s *S) TestTeamServiceRemoveWithApps(c *check.C) {
	teamName := "atreides"
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnDelete: func(t authTypes.Team) error {
				c.Fail()
				return nil
			},
		},
	}

	err := s.conn.Apps().Insert(bson.M{"name": "leto", "teams": []string{teamName}})
	c.Assert(err, check.IsNil)
	err = ts.Remove(context.TODO(), teamName)
	c.Assert(err, check.ErrorMatches, "Apps: leto")
}

func (s *S) TestTeamServiceRemoveWithServiceInstances(c *check.C) {
	teamName := "harkonnen"
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnDelete: func(t authTypes.Team) error {
				c.Fail()
				return nil
			},
		},
	}

	err := s.conn.ServiceInstances().Insert(bson.M{"name": "vladimir", "teams": []string{teamName}})
	c.Assert(err, check.IsNil)
	err = ts.Remove(context.TODO(), teamName)
	c.Assert(err, check.ErrorMatches, "Service instances: vladimir")
}

func (s *S) TestTeamServiceList(c *check.C) {
	teams := []authTypes.Team{
		{Name: "corrino"},
		{Name: "fenring"},
	}
	ts := &teamService{
		storage: &authTypes.MockTeamStorage{
			OnFindAll: func() ([]authTypes.Team, error) {
				return teams, nil
			},
		},
	}

	result, err := ts.List(context.TODO())
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, teams)
}
