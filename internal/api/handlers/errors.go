package handlers

import (
	"errors"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
)

// writeServiceError turns a service error into the status it deserves.
//
// It exists because roles made the generic 500 wrong: a viewer trying to write
// is not a server fault, and a reader told "internal server error" has no idea
// they simply lack the right. Every handler that can reach a permission check
// routes its errors through here, so a new endpoint gets the mapping by
// default rather than by remembering.
//
// ErrNotFound stays 404 for both "no such thing" and "not yours" — that is the
// no-information-leak rule the access layer enforces, and the status has to
// carry it through unchanged.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrForbidden),
		errors.Is(err, service.ErrLastOwner),
		errors.Is(err, service.ErrPersonalTeam),
		errors.Is(err, service.ErrTeamNotEmpty):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrNoAuth):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrAlreadyMember),
		errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInviteInvalid),
		errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrTeamNameEmpty),
		errors.Is(err, service.ErrNotMember),
		errors.Is(err, service.ErrMutualExclusion),
		errors.Is(err, service.ErrAnswerRequired),
		errors.Is(err, service.ErrDuplicateSectionName),
		errors.Is(err, service.ErrSectionHasNoEntries),
		errors.Is(err, service.ErrQuestionDepthLimit),
		errors.Is(err, service.ErrTextReplaceNotFound),
		errors.Is(err, service.ErrTextReplaceOnBlocks):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
