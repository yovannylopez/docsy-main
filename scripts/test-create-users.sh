#!/bin/bash

echo "=== Testing Create Users Endpoint ==="

# Test 1: Crear un usuario individual
echo "Test 1: Crear un usuario individual"
curl -X POST "http://localhost:8100/api/v1/users" \
     -H "Content-Type: application/json" \
     -d '{
       "email": "test@example.com",
       "password": "password123",
       "primer_nombre": "Juan",
       "segundo_nombre": "Pérez",
       "nombre_usuario": "jperez",
       "numero_identificacion": "12345678",
       "tipo_identificacion": "cc",
       "telefono": "3001234567",
       "esta_activo": true,
       "esta_verificado": false,
       "mfa_habilitado": false,
       "nombre_rol": "usuario"
     }' | jq .

echo -e "\n\n"

# Test 2: Crear múltiples usuarios
echo "Test 2: Crear múltiples usuarios"
curl -X POST "http://localhost:8100/api/v1/users" \
     -H "Content-Type: application/json" \
     -d '{
       "usuarios": [
         {
           "email": "user1@example.com",
           "password": "password123",
           "primer_nombre": "María",
           "segundo_nombre": "García",
           "nombre_usuario": "mgarcia",
           "numero_identificacion": "87654321",
           "tipo_identificacion": "cc",
           "telefono": "3007654321",
           "esta_activo": true,
           "esta_verificado": true,
           "mfa_habilitado": false,
           "nombre_rol": "usuario"
         },
         {
           "email": "user2@example.com",
           "password": "password123",
           "primer_nombre": "Carlos",
           "segundo_nombre": "López",
           "nombre_usuario": "clopez",
           "numero_identificacion": "11223344",
           "tipo_identificacion": "ce",
           "telefono": "3001122334",
           "esta_activo": true,
           "esta_verificado": false,
           "mfa_habilitado": true,
           "nombre_rol": "funcionario"
         }
       ]
     }' | jq .

echo -e "\n\n"

# Test 3: Verificar que los usuarios se crearon
echo "Test 3: Verificar usuarios creados"
curl -X GET "http://localhost:8100/api/v1/users?limit=10&offset=0" \
     -H "Content-Type: application/json" | jq .

echo -e "\n\nTest completed!"
