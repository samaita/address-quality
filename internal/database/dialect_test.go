// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package database

import "testing"

func TestRebind(t *testing.T) {
	pg := &sqlDB{pg: true}
	if got, want := pg.rebind(`SELECT a FROM t WHERE b = ? AND c = ?`), `SELECT a FROM t WHERE b = $1 AND c = $2`; got != want {
		t.Fatalf("pg rebind = %q, want %q", got, want)
	}

	lite := &sqlDB{pg: false}
	q := `SELECT a FROM t WHERE b = ?`
	if got := lite.rebind(q); got != q {
		t.Fatalf("sqlite rebind = %q, want unchanged %q", got, q)
	}
}
