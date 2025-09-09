# Dockerfile para MySQL con inicialización automática de la base de datos
FROM mysql:8.0

# Copia el archivo SQL de inicialización
COPY atm.sql /docker-entrypoint-initdb.d/atm.sql

# Exponer el puerto estándar de MySQL
EXPOSE 3306

# Las credenciales y nombre de base de datos se configuran por variables de entorno en Coolify o docker-compose

