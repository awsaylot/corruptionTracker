package graph

import (
	"bytes"
	"clank/internal/models"
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestGetAllNodes(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful retrieval",
			mockSetup: func(db *testutil.MockDB) {
				entity1 := testutil.MockEntity("person", "John Doe")
				entity2 := testutil.MockEntity("company", "Acme Corp")
				db.Entities[entity1.ID] = entity1
				db.Entities[entity2.ID] = entity2
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []models.ExtractedEntity
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, 2)
			},
		},
		{
			name: "database error",
			mockSetup: func(db *testutil.MockDB) {
				db.RetrieveError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupTestRouter()
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			// Register handler
			r.GET("/nodes", func(c *gin.Context) {
				c.Set("db", mockDB)
				GetAllNodes(c)
			})

			// Make request
			req := httptest.NewRequest(http.MethodGet, "/nodes", nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestCreateNode(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    *models.ExtractedEntity
		mockSetup      func(*testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: &models.ExtractedEntity{
				Type: "person",
				Name: "John Doe",
				Properties: map[string]interface{}{
					"title": "CEO",
				},
			},
			mockSetup:      func(db *testutil.MockDB) {},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response models.ExtractedEntity
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "John Doe", response.Name)
			},
		},
		{
			name:           "invalid request",
			requestBody:    &models.ExtractedEntity{},
			mockSetup:      func(db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupTestRouter()
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			// Register handler
			r.POST("/node", func(c *gin.Context) {
				c.Set("db", mockDB)
				CreateNode(c)
			})

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/node", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Make request
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestGetNode(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		mockSetup      func(*testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "node exists",
			nodeID: "test-id",
			mockSetup: func(db *testutil.MockDB) {
				entity := testutil.MockEntity("person", "John Doe")
				entity.ID = "test-id"
				db.Entities[entity.ID] = entity
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response models.ExtractedEntity
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "test-id", response.ID)
				assert.Equal(t, "John Doe", response.Name)
			},
		},
		{
			name:           "node not found",
			nodeID:         "non-existent",
			mockSetup:      func(db *testutil.MockDB) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupTestRouter()
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			// Register handler
			r.GET("/node/:id", func(c *gin.Context) {
				c.Set("db", mockDB)
				GetNode(c)
			})

			// Make request
			req := httptest.NewRequest(http.MethodGet, "/node/"+tt.nodeID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestCreateRelationship(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    *models.ExtractedRelationship
		mockSetup      func(*testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: &models.ExtractedRelationship{
				Type:   "OWNS",
				FromID: "entity1",
				ToID:   "entity2",
				Properties: map[string]interface{}{
					"since": "2025",
				},
			},
			mockSetup: func(db *testutil.MockDB) {
				entity1 := testutil.MockEntity("person", "John Doe")
				entity2 := testutil.MockEntity("company", "Acme Corp")
				entity1.ID = "entity1"
				entity2.ID = "entity2"
				db.Entities[entity1.ID] = entity1
				db.Entities[entity2.ID] = entity2
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response models.ExtractedRelationship
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "OWNS", response.Type)
			},
		},
		{
			name: "invalid entity references",
			requestBody: &models.ExtractedRelationship{
				Type:   "OWNS",
				FromID: "non-existent1",
				ToID:   "non-existent2",
			},
			mockSetup:      func(db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupTestRouter()
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			// Register handler
			r.POST("/relationship", func(c *gin.Context) {
				c.Set("db", mockDB)
				CreateRelationship(c)
			})

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/relationship", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Make request
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
