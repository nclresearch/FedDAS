-- Init SQL statements when starting the database
CREATE DATABASE IF NOT EXISTS agents;
USE agents;
CREATE TABLE IF NOT EXISTS registry
(
    AgentId    CHAR(22) PRIMARY KEY NOT NULL,
    InternalIP CHAR(15)             NOT NULL,
    Prometheus BOOL                 NOT NULL DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS devSupport
(
    ID      INT AUTO_INCREMENT PRIMARY KEY NOT NULL,
    AgentId CHAR(22)                       NOT NULL,
    Device  VARCHAR(32)                    NOT NULL,
    FOREIGN KEY (AgentId) REFERENCES registry (AgentId)
);
CREATE TABLE IF NOT EXISTS sensorSupport
(
    ID      INT AUTO_INCREMENT PRIMARY KEY NOT NULL,
    AgentId CHAR(22)                       NOT NULL,
    Sensor  VARCHAR(32)                    NOT NULL,
    FOREIGN KEY (AgentId) REFERENCES registry (AgentId)
);
CREATE TABLE IF NOT EXISTS containers
(
    ID          INT AUTO_INCREMENT PRIMARY KEY NOT NULL,
    ContainerId CHAR(64)                       NOT NULL,
    AgentId     CHAR(22)                       NOT NULL,
    Status      VARCHAR(10)                    NOT NULL,
    FOREIGN KEY (AgentId) REFERENCES registry (AgentId)
);