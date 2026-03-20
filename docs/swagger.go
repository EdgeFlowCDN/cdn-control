package docs

// OpenAPISpec is a hand-written OpenAPI 3.0 specification covering the main
// EdgeFlow control plane API endpoints.
const OpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "EdgeFlow CDN Control Plane API",
    "description": "REST API for managing CDN domains, origins, cache rules, purge operations, certificates, and edge nodes.",
    "version": "1.0.0",
    "contact": {
      "name": "EdgeFlow Team"
    }
  },
  "servers": [
    {
      "url": "/",
      "description": "Current server"
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "LoginRequest": {
        "type": "object",
        "required": ["username", "password"],
        "properties": {
          "username": { "type": "string" },
          "password": { "type": "string" }
        }
      },
      "LoginResponse": {
        "type": "object",
        "properties": {
          "token": { "type": "string" }
        }
      },
      "Domain": {
        "type": "object",
        "properties": {
          "id":         { "type": "string", "format": "uuid" },
          "domain":     { "type": "string" },
          "status":     { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" }
        }
      },
      "CreateDomainRequest": {
        "type": "object",
        "required": ["domain"],
        "properties": {
          "domain": { "type": "string" }
        }
      },
      "Origin": {
        "type": "object",
        "properties": {
          "id":        { "type": "string", "format": "uuid" },
          "domain_id": { "type": "string", "format": "uuid" },
          "address":   { "type": "string" },
          "weight":    { "type": "integer" },
          "active":    { "type": "boolean" }
        }
      },
      "CreateOriginRequest": {
        "type": "object",
        "required": ["address"],
        "properties": {
          "address": { "type": "string" },
          "weight":  { "type": "integer" }
        }
      },
      "PurgeURLRequest": {
        "type": "object",
        "required": ["domain", "targets"],
        "properties": {
          "domain":  { "type": "string" },
          "targets": { "type": "array", "items": { "type": "string" } }
        }
      },
      "PurgeAllRequest": {
        "type": "object",
        "required": ["domain"],
        "properties": {
          "domain": { "type": "string" }
        }
      },
      "Node": {
        "type": "object",
        "properties": {
          "id":             { "type": "string", "format": "uuid" },
          "hostname":       { "type": "string" },
          "region":         { "type": "string" },
          "status":         { "type": "string" },
          "last_heartbeat": { "type": "string", "format": "date-time" }
        }
      },
      "Error": {
        "type": "object",
        "properties": {
          "error": { "type": "string" }
        }
      }
    }
  },
  "paths": {
    "/api/v1/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Authenticate and obtain a JWT token",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/LoginRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful login",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/LoginResponse" }
              }
            }
          },
          "401": {
            "description": "Invalid credentials",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/api/v1/domains": {
      "get": {
        "tags": ["Domains"],
        "summary": "List all domains",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "List of domains",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Domain" }
                }
              }
            }
          }
        }
      },
      "post": {
        "tags": ["Domains"],
        "summary": "Create a new domain",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateDomainRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Domain created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Domain" }
              }
            }
          }
        }
      }
    },
    "/api/v1/domains/{id}": {
      "get": {
        "tags": ["Domains"],
        "summary": "Get a domain by ID",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": {
            "description": "Domain details",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Domain" }
              }
            }
          }
        }
      },
      "put": {
        "tags": ["Domains"],
        "summary": "Update a domain",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateDomainRequest" }
            }
          }
        },
        "responses": {
          "200": { "description": "Domain updated" }
        }
      },
      "delete": {
        "tags": ["Domains"],
        "summary": "Delete a domain",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Domain deleted" }
        }
      }
    },
    "/api/v1/domains/{id}/origins": {
      "get": {
        "tags": ["Origins"],
        "summary": "List origins for a domain",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": {
            "description": "List of origins",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Origin" }
                }
              }
            }
          }
        }
      },
      "post": {
        "tags": ["Origins"],
        "summary": "Add an origin to a domain",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateOriginRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Origin created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Origin" }
              }
            }
          }
        }
      }
    },
    "/api/v1/purge/url": {
      "post": {
        "tags": ["Purge"],
        "summary": "Purge specific URLs from cache",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/PurgeURLRequest" }
            }
          }
        },
        "responses": {
          "200": { "description": "Purge task created" }
        }
      }
    },
    "/api/v1/purge/all": {
      "post": {
        "tags": ["Purge"],
        "summary": "Purge all cached content for a domain",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/PurgeAllRequest" }
            }
          }
        },
        "responses": {
          "200": { "description": "Full purge task created" }
        }
      }
    },
    "/api/v1/nodes": {
      "get": {
        "tags": ["Nodes"],
        "summary": "List all edge nodes",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "List of nodes",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Node" }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/nodes/{id}": {
      "get": {
        "tags": ["Nodes"],
        "summary": "Get a node by ID",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": {
            "description": "Node details",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Node" }
              }
            }
          }
        }
      }
    }
  }
}`
