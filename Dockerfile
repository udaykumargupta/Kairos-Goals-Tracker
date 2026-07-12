# syntax=docker/dockerfile:1

# ---- Build stage: compile the Spring Boot app (also bundles frontend/ into static) ----
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /workspace

# The backend pom copies ../frontend into the app's static content, so both are needed.
COPY backend/pom.xml backend/pom.xml
COPY backend/src backend/src
COPY frontend frontend

RUN mvn -f backend/pom.xml -B -DskipTests clean package

# ---- Run stage: small JRE image running the fat jar ----
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /workspace/backend/target/*.jar app.jar

# Tuned for small (512 MB) free instances.
ENV JAVA_OPTS="-XX:MaxRAMPercentage=70 -XX:+UseSerialGC"

# Render/most PaaS inject $PORT; the app reads it via server.port=${PORT:8080}.
EXPOSE 8080
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
