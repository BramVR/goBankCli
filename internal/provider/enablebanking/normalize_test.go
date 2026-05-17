package enablebanking

import "testing"

func TestParseInstitutionIDRequiresCountryAndName(t *testing.T) {
	country, name, err := ParseInstitutionID(" be : Belfius ")
	if err != nil {
		t.Fatal(err)
	}
	if country != "BE" || name != "Belfius" {
		t.Fatalf("country=%q name=%q", country, name)
	}
	if _, _, err := ParseInstitutionID("Belfius"); err != ErrInvalidInstitutionID {
		t.Fatalf("err = %v, want invalid institution id", err)
	}
}

func TestNormalizeAccountFallsBackWhenIdentificationHashMissing(t *testing.T) {
	session := sessionPayload{
		SessionID: "session-1",
		ASPSP:     aspspPayload{Name: "Belfius", Country: "BE"},
	}
	account, err := NormalizeAccount(session, accountPayload{
		UID: "uid-1",
		AccountID: accountIdentification{
			Other: otherAccountIdentification{
				SchemeName:     "BBAN",
				Identification: "123",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.ProviderAccountID != "BBAN:123" || account.ProviderResourceID != "uid-1" {
		t.Fatalf("account = %+v", account)
	}
}

func TestNormalizeAccountRejectsUIDOnlyIdentity(t *testing.T) {
	_, err := NormalizeAccount(sessionPayload{}, accountPayload{UID: "session-uid"})
	if err != ErrMissingStableAccountID {
		t.Fatalf("err = %v, want missing stable account id", err)
	}
}

func TestInstitutionIDEmptyWhenMissingParts(t *testing.T) {
	if got := InstitutionID("", ""); got != "" {
		t.Fatalf("institution id = %q, want empty", got)
	}
	if got := InstitutionID("BE", "Belfius"); got != "BE:Belfius" {
		t.Fatalf("institution id = %q", got)
	}
}
