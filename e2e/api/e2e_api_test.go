//go:build api

package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/verzac/grocer-discord-bot/dto"
	"github.com/verzac/grocer-discord-bot/e2e/api/apiharness"
	"github.com/verzac/grocer-discord-bot/models"
)

var apiSess *apiharness.APITestSession

func TestMain(m *testing.M) {
	apiSess = apiharness.SetupAPI()
	if err := apiSess.CleanupAllGroceries(); err != nil {
		panic(err)
	}
	code := m.Run()
	if err := apiSess.CleanupAllGroceries(); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func cleanupGroceries(t *testing.T) {
	t.Helper()
	require.NoError(t, apiSess.CleanupAllGroceries())
}

func postGroceries(t *testing.T, body string) models.GroceryEntry {
	t.Helper()
	res, err := apiSess.PostJSON("/groceries", []byte(body))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", string(b))
	var out models.GroceryEntry
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func TestGetGroceryListsEmpty(t *testing.T) {
	cleanupGroceries(t)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Empty(t, lists.GroceryEntries)

	t.Cleanup(func() { _ = apiSess.CleanupAllGroceries() })
}

func TestCreateGroceryEntry(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	e := postGroceries(t, `{"item_desc":"Chicken"}`)
	require.NotZero(t, e.ID)
	require.Equal(t, "Chicken", e.ItemDesc)
}

func TestCreateAndReadBack(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	postGroceries(t, `{"item_desc":"api-e2e-readback"}`)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	found := false
	for _, g := range lists.GroceryEntries {
		if g.ItemDesc == "api-e2e-readback" {
			found = true
			break
		}
	}
	require.True(t, found, "expected grocery entry in grocery_entries")
}

func TestDeleteSingleGroceryEntry(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	e := postGroceries(t, `{"item_desc":"api-e2e-single-del"}`)

	res, err := apiSess.DeleteNoBody("/groceries/" + uintToStr(e.ID))
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	for _, g := range lists.GroceryEntries {
		require.NotEqual(t, e.ID, g.ID, "entry should be deleted")
	}
}

func TestBatchDeleteGroceries(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	a := postGroceries(t, `{"item_desc":"api-e2e-batch-a"}`)
	b := postGroceries(t, `{"item_desc":"api-e2e-batch-b"}`)
	c := postGroceries(t, `{"item_desc":"api-e2e-batch-c"}`)

	payload, err := json.Marshal(dto.GroceryBatchDeleteRequest{IDs: []uint{a.ID, b.ID}})
	require.NoError(t, err)
	res, err := apiSess.DeleteJSON("/groceries", payload)
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, lists.GroceryEntries, 1)
	require.Equal(t, c.ID, lists.GroceryEntries[0].ID)
}

func TestDeleteNonExistentEntry(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	res, err := apiSess.DeleteNoBody("/groceries/999999999")
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestBatchDeleteNonExistentEntry(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	e := postGroceries(t, `{"item_desc":"api-e2e-batch-404"}`)
	payload, err := json.Marshal(dto.GroceryBatchDeleteRequest{IDs: []uint{e.ID, 999999998}})
	require.NoError(t, err)
	res, err := apiSess.DeleteJSON("/groceries", payload)
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	found := false
	for _, g := range lists.GroceryEntries {
		if g.ID == e.ID {
			found = true
			break
		}
	}
	require.True(t, found, "batch delete with missing id must not delete existing rows")
}

func TestCreateValidationErrors(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	res, err := apiSess.PostJSON("/groceries", []byte(`{}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))

	res2, err := apiSess.PostJSON("/groceries", []byte(`{"id":1,"item_desc":"x"}`))
	require.NoError(t, err)
	b2, err := apiharness.ReadBodyAndClose(res2)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res2.StatusCode, "%s", string(b2))
}

func TestGetRegistrations(t *testing.T) {
	res, err := apiSess.Get("/registrations")
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode, "%s", string(b))
	var regs []json.RawMessage
	require.NoError(t, json.Unmarshal(b, &regs))
	require.NotNil(t, regs)
}

func TestCRDFlow(t *testing.T) {
	cleanupGroceries(t)
	defer cleanupGroceries(t)

	postGroceries(t, `{"item_desc":"api-e2e-crdflow-1"}`)
	postGroceries(t, `{"item_desc":"api-e2e-crdflow-2"}`)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, lists.GroceryEntries, 2)

	var idToDelete uint
	for _, g := range lists.GroceryEntries {
		if g.ItemDesc == "api-e2e-crdflow-1" {
			idToDelete = g.ID
			break
		}
	}
	require.NotZero(t, idToDelete)

	res, err := apiSess.DeleteNoBody("/groceries/" + uintToStr(idToDelete))
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	lists, status, err = apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, lists.GroceryEntries, 1)

	restID := lists.GroceryEntries[0].ID
	payload, err := json.Marshal(dto.GroceryBatchDeleteRequest{IDs: []uint{restID}})
	require.NoError(t, err)
	res2, err := apiSess.DeleteJSON("/groceries", payload)
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res2)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res2.StatusCode)

	lists, status, err = apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Empty(t, lists.GroceryEntries)
}

func cleanupGroceryLists(t *testing.T) {
	t.Helper()
	require.NoError(t, apiSess.CleanupAllGroceryLists())
}

func postGroceryList(t *testing.T, body string) models.GroceryList {
	t.Helper()
	res, err := apiSess.PostGroceryList([]byte(body))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", string(b))
	var out models.GroceryList
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func TestCreateGroceryList(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"fruit"}`)
	require.NotZero(t, gl.ID)
	require.Equal(t, "fruit", gl.ListLabel)
}

func TestCreateGroceryListReadBack(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	postGroceryList(t, `{"list_label":"veggies"}`)

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	found := false
	for _, gl := range lists.GroceryLists {
		if gl.ListLabel == "veggies" {
			found = true
			break
		}
	}
	require.True(t, found, "expected grocery list in grocery_lists")
}

func TestCreateGroceryListDuplicate(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	postGroceryList(t, `{"list_label":"fruit"}`)

	res, err := apiSess.PostGroceryList([]byte(`{"list_label":"fruit"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))
}

func TestCreateGroceryListInvalidLabel(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	res, err := apiSess.PostGroceryList([]byte(`{"list_label":"fruit123"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))
}

func TestCreateGroceryListLimitReached(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	// Match handlers/grolist.go: the existing command path is the source of truth.
	postGroceryList(t, `{"list_label":"listA"}`)
	postGroceryList(t, `{"list_label":"listB"}`)

	res, err := apiSess.PostGroceryList([]byte(`{"list_label":"listC"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))
}

func TestCreateGroceryListMissingLabel(t *testing.T) {
	res, err := apiSess.PostGroceryList([]byte(`{}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))
}

func TestDeleteGroceryList(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"todelete"}`)

	res, err := apiSess.DeleteGroceryList(gl.ID)
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode, "%s", string(b))

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	for _, l := range lists.GroceryLists {
		require.NotEqual(t, gl.ID, l.ID, "list should be deleted")
	}
}

func TestDeleteGroceryListNotFound(t *testing.T) {
	res, err := apiSess.DeleteGroceryList(999999999)
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode, "%s", string(b))
}

func TestDeleteGroceryListZeroID(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"zerodelete"}`)

	res, err := apiSess.DeleteGroceryList(0)
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	found := false
	for _, l := range lists.GroceryLists {
		if l.ID == gl.ID {
			found = true
			break
		}
	}
	require.True(t, found, "zero ID delete must not delete an existing list")
}

func TestPatchGroceryList(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"patchme"}`)

	res, err := apiSess.PatchGroceryList(gl.ID, []byte(`{"fancy_name":"My Patched List"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode, "%s", string(b))

	var updated models.GroceryList
	require.NoError(t, json.Unmarshal(b, &updated))
	require.NotNil(t, updated.FancyName)
	require.Equal(t, "My Patched List", *updated.FancyName)
}

func TestPatchGroceryListClearFancyName(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"clearname","fancy_name":"Clear Me"}`)

	res, err := apiSess.PatchGroceryList(gl.ID, []byte(`{"fancy_name":null}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode, "%s", string(b))

	var updated models.GroceryList
	require.NoError(t, json.Unmarshal(b, &updated))
	require.Nil(t, updated.FancyName)
}

func TestPatchGroceryListNotFound(t *testing.T) {
	res, err := apiSess.PatchGroceryList(999999999, []byte(`{"fancy_name":"Ghost"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode, "%s", string(b))
}

func TestPatchGroceryListZeroID(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"zeropatch"}`)

	res, err := apiSess.PatchGroceryList(0, []byte(`{"fancy_name":"Should Not Patch"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))

	lists, status, err := apiSess.FetchGroceryLists()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	for _, l := range lists.GroceryLists {
		if l.ID == gl.ID {
			require.Nil(t, l.FancyName)
			return
		}
	}
	require.Fail(t, "expected list to remain after zero ID patch")
}

func TestPatchGroceryListMissingFancyName(t *testing.T) {
	cleanupGroceryLists(t)
	defer cleanupGroceryLists(t)

	gl := postGroceryList(t, `{"list_label":"patchvalidate"}`)

	res, err := apiSess.PatchGroceryList(gl.ID, []byte(`{}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", string(b))
}

func TestDeleteGroceryListWithEntries(t *testing.T) {
	cleanupGroceries(t)
	cleanupGroceryLists(t)
	defer func() {
		cleanupGroceries(t)
		cleanupGroceryLists(t)
	}()

	gl := postGroceryList(t, `{"list_label":"nonempty"}`)

	glID := gl.ID
	body, err := json.Marshal(map[string]interface{}{
		"item_desc":       "some item",
		"grocery_list_id": glID,
	})
	require.NoError(t, err)
	res, err := apiSess.PostJSON("/groceries", body)
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", string(b))

	res, err = apiSess.DeleteGroceryList(gl.ID)
	require.NoError(t, err)
	b, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, res.StatusCode, "%s", string(b))
}

func uintToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// Intent token cases need a real Bearer user: set E2E_API_BEARER_TOKEN and E2E_API_GUILD_ID.
// TestIntentTokenRevokedOnLogout also needs E2E_API_BEARER_ALLOW_LOGOUT=1 because it deletes that user's session.

func bearerSessionOrSkip(t *testing.T) *apiharness.APITestSession {
	t.Helper()
	s := apiharness.SetupBearerAPI()
	if s == nil {
		t.Skip("E2E_API_BEARER_TOKEN and E2E_API_GUILD_ID are required for intent token tests")
	}
	return s
}

func mintIntentToken(t *testing.T, s *apiharness.APITestSession) string {
	t.Helper()
	res, err := s.PostNoBody("/auth/intent-token")
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", string(b))
	var out struct {
		IntentToken string `json:"intent_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	require.NoError(t, json.Unmarshal(b, &out))
	require.True(t, strings.HasPrefix(out.IntentToken, "gbi_"), out.IntentToken)
	require.Equal(t, int64(2592000), out.ExpiresIn)
	return out.IntentToken
}

func intentStatus(t *testing.T, s *apiharness.APITestSession, method, path string) int {
	t.Helper()
	var res *http.Response
	var err error
	switch method {
	case http.MethodGet:
		res, err = s.Get(path)
	case http.MethodPost:
		res, err = s.PostJSON(path, []byte(`{"item_desc":"api-e2e-intent-rejected"}`))
	default:
		t.Fatalf("unsupported method %s", method)
	}
	require.NoError(t, err)
	_, err = apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	return res.StatusCode
}

func TestIntentTokenCreateGrocery(t *testing.T) {
	bearer := bearerSessionOrSkip(t)
	intent := bearer.WithAuthorization("Bearer " + mintIntentToken(t, bearer))

	res, err := intent.PostJSON("/groceries", []byte(`{"item_desc":"api-e2e-intent-token"}`))
	require.NoError(t, err)
	b, err := apiharness.ReadBodyAndClose(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode, "%s", string(b))
	var e models.GroceryEntry
	require.NoError(t, json.Unmarshal(b, &e))
	require.Equal(t, "api-e2e-intent-token", e.ItemDesc)

	res, err = bearer.DeleteNoBody("/groceries/" + uintToStr(e.ID))
	require.NoError(t, err)
	_, _ = apiharness.ReadBodyAndClose(res)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestIntentTokenForbiddenElsewhere(t *testing.T) {
	bearer := bearerSessionOrSkip(t)
	intent := bearer.WithAuthorization("Bearer " + mintIntentToken(t, bearer))

	require.Equal(t, http.StatusForbidden, intentStatus(t, intent, http.MethodGet, "/grocery-lists"))
	require.Equal(t, http.StatusForbidden, intentStatus(t, intent, http.MethodPost, "/auth/intent-token"))
}

func TestIntentTokenGarbage(t *testing.T) {
	bearer := bearerSessionOrSkip(t)
	intent := bearer.WithAuthorization("Bearer gbi_not-a-real-token")

	require.Equal(t, http.StatusUnauthorized, intentStatus(t, intent, http.MethodPost, "/groceries"))
}

func TestIntentTokenRevokedOnRemint(t *testing.T) {
	bearer := bearerSessionOrSkip(t)
	old := bearer.WithAuthorization("Bearer " + mintIntentToken(t, bearer))
	_ = mintIntentToken(t, bearer)

	require.Equal(t, http.StatusUnauthorized, intentStatus(t, old, http.MethodPost, "/groceries"))
}

func TestIntentTokenRevokedOnLogout(t *testing.T) {
	bearer := bearerSessionOrSkip(t)
	if os.Getenv("E2E_API_BEARER_ALLOW_LOGOUT") != "1" {
		t.Skip("set E2E_API_BEARER_ALLOW_LOGOUT=1 to run; this logs the E2E user out")
	}
	intent := bearer.WithAuthorization("Bearer " + mintIntentToken(t, bearer))

	res, err := bearer.PostNoBody("/auth/logout")
	require.NoError(t, err)
	_, _ = apiharness.ReadBodyAndClose(res)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	require.Equal(t, http.StatusUnauthorized, intentStatus(t, intent, http.MethodPost, "/groceries"))
}
