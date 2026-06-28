package api

import "github.com/go-chi/chi/v5"

func (s *Server) profileRoutes(r chi.Router) {
	r.Get("/", s.handleMe)
	r.Put("/", s.handleUpdateProfile)
	r.Delete("/", s.handleDeleteAccount)
	r.Put("/password", s.handleChangePassword)
}

func (s *Server) exerciseRoutes(r chi.Router) {
	r.Get("/", s.handleListExercises)
	r.Post("/", s.handleCreateExercise)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", s.handleGetExercise)
		r.Put("/", s.handleUpdateExercise)
		r.Post("/archive", s.handleArchiveExercise(true))
		r.Post("/restore", s.handleArchiveExercise(false))
		r.Get("/records", s.handleExerciseRecords)
		r.Get("/progress", s.handleExerciseProgress)
		r.Get("/last", s.handleLastSession)
		r.Get("/notes", s.handleGetExerciseNotes)
		r.Put("/notes", s.handlePutExerciseNotes)
	})
}

func (s *Server) templateRoutes(r chi.Router) {
	r.Get("/", s.handleListTemplates)
	r.Post("/", s.handleCreateTemplate)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", s.handleGetTemplate)
		r.Put("/", s.handleUpdateTemplate)
		r.Delete("/", s.handleDeleteTemplate)
		r.Post("/start", s.handleStartTemplate)
		r.Post("/copy", s.handleCopyTemplate)
	})
}

func (s *Server) programRoutes(r chi.Router) {
	r.Get("/", s.handleListPrograms)
	r.Post("/", s.handleCreateProgram)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", s.handleGetProgram)
		r.Put("/", s.handleUpdateProgram)
		r.Delete("/", s.handleDeleteProgram)
		r.Post("/activate", s.handleActivateProgram)
		r.Post("/start", s.handleStartProgram)
	})
}

func (s *Server) workoutRoutes(r chi.Router) {
	r.Get("/", s.handleListWorkouts)
	r.Post("/", s.handleCreateWorkout)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", s.handleGetWorkout)
		r.Put("/", s.handleUpdateWorkout)
		r.Delete("/", s.handleDeleteWorkout)
		r.Post("/finish", s.handleFinishWorkout)
		r.Post("/repeat", s.handleRepeatWorkout)
		r.Put("/order", s.handleReorderSets)
		r.Post("/warmups", s.handleAddWarmups)
		r.Post("/sets", s.handleAddSet)
		r.Put("/sets/{setID}", s.handleUpdateSet)
		r.Delete("/sets/{setID}", s.handleDeleteSet)
	})
}
