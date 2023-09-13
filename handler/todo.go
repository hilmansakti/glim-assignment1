package handler

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/PrinceNorin/todo-go/types"
	echo "github.com/labstack/echo/v4"
)

var (
	errBadRequest     = newHTTPError(http.StatusBadRequest, "bad request")
	errRecordNotFound = newHTTPError(http.StatusNotFound, "record not found")
)

type Handler struct {
	m      sync.Mutex
	lastID int
	todos  map[int]*types.Todo
}

func NewHandler() *Handler {
	return &Handler{
		todos: make(map[int]*types.Todo),
	}
}

// CreateTodo godoc
// @Summary Create a todo
// @Description Create a new todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body types.Todo true "New Todo"
// @Success 201 {object} types.Todo
// @Failure 400 {object} HTTPError
// @Router /todos [post]
func (h *Handler) CreateTodo(c echo.Context) error {
	return h.withLockContext(func() error {
		var todo types.Todo
		if err := c.Bind(&todo); err != nil {
			return errBadRequest
		}

		if todo.Name == "" {
			return errBadRequest
		}

		h.lastID++
		todo.ID = h.lastID
		h.todos[todo.ID] = &todo
		return c.JSON(http.StatusCreated, &todo)
	})
}

// FindTodos godoc
// @Summary Get all todos
// @Description Get all todo items
// @Tags todos
// @Accept json
// @Produce json
// @Success 200 {array} types.Todo
// @Failure 404 {object} HTTPError
// @Router /todos [get]
func (h *Handler) FindTodos(c echo.Context) error {
	todos := make([]*types.Todo, 0)
	for _, todo := range h.todos {
		todos = append(todos, todo)
	}

	return c.JSON(http.StatusOK, todos)
}

// GetTodo godoc
// @Summary Get a todo
// @Description Get a todo item
// @Tags todos
// @ID get-string-by-int
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} types.Todo
// @Failure 404 {object} HTTPError
// @Router /todos/{id} [get]
func (h *Handler) GetTodo(c echo.Context) error {
	id, err := getTodoID(c)
	if err != nil {
		return err
	}

	todo, ok := h.todos[id]
	if !ok {
		return errRecordNotFound
	}

	return c.JSON(http.StatusOK, todo)
}

// UpdateTodo godoc
// @Summary Update a todo
// @Description Update a todo item
// @Tags todos
// @ID get-string-by-int
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body types.Todo true "Update Todo"
// @Success 201 {object} types.Todo
// @Failure 400 {object} HTTPError
// @Router /todos/{id} [PUT]
func (h *Handler) UpdateTodo(c echo.Context) error {
	return h.withLockContext(func() error {
		var todo types.Todo
		if err := c.Bind(&todo); err != nil {
			return errBadRequest
		}

		id, err := getTodoID(c)
		if err != nil {
			return err
		}

		_, ok := h.todos[id]
		if !ok {
			return errRecordNotFound
		}

		//updating
		h.todos[id].Name = todo.Name
		h.todos[id].Completed = todo.Completed

		return c.JSON(http.StatusOK, h.todos[id])
	})
}

// DeleteTodo godoc
// @Summary Delete a todo
// @Description Delete a todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204
// @Failure 400 {object} HTTPError
// @Router /todos/:id [delete]
func (h *Handler) DeleteTodo(c echo.Context) error {
	return h.withLockContext(func() error {
		id, err := getTodoID(c)
		if err != nil {
			return err
		}

		_, ok := h.todos[id]
		if !ok {
			return errRecordNotFound
		}

		delete(h.todos, id)

		return c.JSON(http.StatusNoContent, nil)
	})
}

func (h *Handler) withLockContext(fn func() error) error {
	h.m.Lock()
	defer h.m.Unlock()

	return fn()
}

func getTodoID(c echo.Context) (int, error) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)

	if err != nil {
		return 0, errRecordNotFound
	}

	return id, nil
}

func newHTTPError(code int, msg string) *HTTPError {
	return &HTTPError{code: code, msg: msg}
}
