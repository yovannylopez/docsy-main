#!/bin/bash
echo 'Iniciando servidor en modo de prueba...'
echo 'El servidor estará disponible en: http://localhost:8100'
echo 'Archivos estáticos disponibles en: http://localhost:8100/uploads/'
echo 'Ejemplo: http://localhost:8100/uploads/2025/10/01/1759295687297720107_3290_manual-vut--sige-2%20(1).pdf'
echo ''
echo 'Para probar CORS desde Angular (http://localhost:4200), puedes usar:'
echo 'fetch("http://localhost:8100/uploads/2025/10/01/1759295687297720107_3290_manual-vut--sige-2%20(1).pdf")'
echo ''
echo 'Presiona Ctrl+C para detener el servidor'
echo '========================================'
./bin/docsy-main
