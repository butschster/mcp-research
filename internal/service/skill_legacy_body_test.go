package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
)

func TestSkill_MetadataRepairPreservesHistoricalBody(t *testing.T) {
	for _, historical := range []struct {
		name, body       string
		changedBodyError error
	}{
		{"oversized", " \r\n" + strings.Repeat("я", domain.SkillBodyMax+1) + "\r\n ", ErrSkillBodyLong},
		{"whitespace", " \r\n\t ", ErrSkillBodyEmpty},
	} {
		t.Run(historical.name, func(t *testing.T) {
			body := historical.body
			k := newSkillKit(t)
			owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
			sk, err := k.skills.CreatePrivate(owner, research.ID, skillInput("Legacy instruction"))
			if err != nil {
				t.Fatal(err)
			}
			// Migration imports historical bodies without modern editorial caps.
			sk.Body, sk.NeedsTrigger = body, true
			if err := k.repo.Update(owner, sk); err != nil {
				t.Fatal(err)
			}
			in := SkillInput{Name: "Reviewed instruction", Description: "Use when conducting this research.", Body: body}
			updated, err := k.skills.Update(owner, sk.ID, in)
			if err != nil {
				t.Fatalf("repair trigger: %v", err)
			}
			if updated.Body != body || updated.NeedsTrigger {
				t.Fatal("repair changed historical body or left trigger pending")
			}
			stored, err := k.repo.Body(owner, sk.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored != body {
				t.Fatal("stored body was not preserved byte for byte")
			}
			in.Name, in.DescriptionUntouched = "Renamed instruction", true
			if updated, err = k.skills.Update(owner, sk.ID, in); err != nil || updated.Body != body {
				t.Fatalf("rename: %v", err)
			}
			for _, tc := range []struct {
				name, description, body string
				want                    error
			}{
				{"", in.Description, body, ErrSkillNameEmpty},
				{in.Name, "", body, ErrSkillDescriptionEmpty},
				{in.Name, strings.Repeat("я", domain.SkillDescriptionMax+1), body, ErrSkillDescriptionLong},
				{in.Name, in.Description, strings.Repeat("я", domain.SkillBodyMax+2), ErrSkillBodyLong},
				{in.Name, in.Description, body + " \t", historical.changedBodyError},
			} {
				_, err := k.skills.Update(owner, sk.ID, SkillInput{Name: tc.name, Description: tc.description, Body: tc.body})
				if !errors.Is(err, tc.want) {
					t.Errorf("update error = %v, want %v", err, tc.want)
				}
			}
			in.Body = strings.Repeat("я", domain.SkillBodyMax+1)
			if _, err := k.skills.CreatePrivate(owner, research.ID, in); !errors.Is(err, ErrSkillBodyLong) {
				t.Fatalf("create oversized: %v", err)
			}
		})
	}
}
