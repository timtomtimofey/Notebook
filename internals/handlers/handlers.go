package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"notebook/internals/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	ns storage.NoteStorage
}

func NewHandler(ns storage.NoteStorage) *Handler {
	return &Handler{ns: ns}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("List: error parsing {id}"))
		return
	}
	if exist, err := h.ns.IsNote(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("List: storage error in IsNote: " + err.Error()))
		return
	} else if !exist {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("List: note " + id + " not exists"))
		return
	}
	note, err := h.ns.GetNote(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("List: storage error in GetNote: " + err.Error()))
		return
	}
	res, err := json.Marshal(note)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("List: error while marshalling note: " + err.Error()))
		return
	}
	w.Write(res)
}

func (h *Handler) ListRange(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	offset := 0
	limit := -1
	// if parsing results in error then just assumes default values for offset and limit
	if i, err := strconv.ParseInt(vars.Get("offset"), 10, 32); err == nil && i >= 0 {
		offset = int(i)
	}
	if i, err := strconv.ParseInt(vars.Get("limit"), 10, 32); err == nil && i >= 0 {
		limit = int(i)
	}

	notes, err := h.ns.GetRangeNotes(r.Context(), offset, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ListRange: storage error in GetRangeNotes: " + err.Error()))
		return
	}
	res, err := json.Marshal(notes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ListRange: error while marshalling notes: " + err.Error()))
		return
	}
	w.Write(res)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	switch mime := r.Header.Get("content-type"); mime {
	case "application/json":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Add: error while reading request body: " + err.Error()))
			return
		}
		var note storage.Note
		if err := json.Unmarshal(body, &note); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Add: error while unmarshalling body"))
			return
		}
		if ok := note.VerifyNote(); !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Add: note is invalid (perhaps some mandatory field is empty)"))
			return
		}

		if exist, err := h.ns.IsNote(r.Context(), note.ID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Add: storage error in IsNote: " + err.Error()))
			return
		} else if exist {
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("Add: note with ID='%s' already exists", note.ID)
			w.Write([]byte(msg))
			return
		}

		if note.ID == "" {
			note.ID = uuid.NewString()
		}

		// log.Printf("%#v", note)

		res, err := h.ns.AddNote(r.Context(), &note)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Add: storage error in AddNote: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusCreated)
		if raw, err := json.Marshal(res); err != nil {
			w.Write([]byte("Add: error while marshalling note: " + err.Error()))
			return
		} else {
			w.Write(raw)
		}
	case "":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Add: header 'content-type' is unset"))
	default:
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Add: unsupported content-type: '" + mime + "' expected 'application/json'"))
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Update: error parsing {id}"))
		return
	}
	if exist, err := h.ns.IsNote(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Update: storage error in IsNote: " + err.Error()))
		return
	} else if !exist {
		w.WriteHeader(http.StatusNotFound)
		msg := fmt.Sprintf("Update error: note with ID='%s' not exists", id)
		w.Write([]byte(msg))
		return
	}

	switch mime := r.Header.Get("content-type"); mime {
	case "application/json":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Update: error while reading request body: " + err.Error()))
			return
		}
		var note storage.Note
		if err := json.Unmarshal(body, &note); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Update: error while unmarshalling body"))
			return
		}
		res, err := h.ns.UpdateNote(r.Context(), id, &note)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Update: storage error in UpdateNote: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		if raw, err := json.Marshal(res); err != nil {
			w.Write([]byte("Update: error while marshalling note: " + err.Error()))
			return
		} else {
			w.Write(raw)
		}
	case "":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Update: header 'content-type' is unset"))
	default:
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Update: unsupported content-type: '" + mime + "' expected 'application/json'"))
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Delete: error parsing {id}"))
		return
	}
	if ok, err := h.ns.IsNote(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Update: storage error in IsNote: " + err.Error()))
		return
	} else if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.ns.DeleteNote(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Delete: storage error in DeleteNote: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
