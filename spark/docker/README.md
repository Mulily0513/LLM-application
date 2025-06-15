# Hadoop、Spark 和 Hive Docker 环境

此目录包含一个 Docker Compose 配置，用于快速启动一个包含 Hadoop (HDFS)、Spark (Standalone) 和 Hive 的学习环境。

## 先决条件

- Docker
- Docker Compose

## 概述

此配置提供了一个简化的、适合开发和学习目的的大数据技术栈。它包括：

-   **Hadoop 分布式文件系统 (HDFS)**：用于分布式存储。
-   **Apache Spark (独立集群模式)**：用于分布式数据处理。
-   **Apache Hive**: 用于数据仓库和在 HDFS 数据上进行类 SQL 查询，它使用 PostgreSQL 作为其 Metastore 的后端。

此配置通过排除 Hadoop YARN 来保持精简。Hive 将以本地模式运行 MapReduce 作业。

## 服务

`docker-compose.yml` 文件定义了以下服务：

1.  `namenode`:
    *   **镜像**: `bde2020/hadoop-namenode:2.0.0-hadoop3.2.1-java8`
    *   **用途**: Hadoop HDFS NameNode。管理文件系统命名空间和元数据。
    *   **Web UI**: [http://localhost:9870](http://localhost:9870)
    *   **HDFS 访问**: `hdfs://namenode:9000` (Docker 网络内部)
2.  `datanode`:
    *   **镜像**: `bde2020/hadoop-datanode:2.0.0-hadoop3.2.1-java8`
    *   **用途**: Hadoop HDFS DataNode。存储实际的数据块。
3.  `spark-master`:
    *   **镜像**: `bde2020/spark-master:3.0.0-hadoop3.2`
    *   **用途**: Spark Standalone Master。管理 Spark worker 节点。
    *   **Web UI**: [http://localhost:8080](http://localhost:8080)
    *   **Spark URL**: `spark://spark-master:7077` (内部使用及用于 worker 连接)
4.  `spark-worker-1`:
    *   **镜像**: `bde2020/spark-worker:3.0.0-hadoop3.2`
    *   **用途**: Spark Standalone Worker。执行 Spark 任务。
    *   **Web UI**: [http://localhost:8081](http://localhost:8081) (通常，请检查 Spark Master UI 以获取实际的 worker UI 链接)
5.  `hive-metastore-db`:
    *   **镜像**: `postgres:12`
    *   **用途**: PostgreSQL 数据库，用作 Hive Metastore 的后端。
    *   **端口**: `5432` (如果需要，可用于直接访问)
6.  `hive-metastore`:
    *   **镜像**: `bde2020/hive-metastore-postgresql:2.3.0`
    *   **用途**: Hive Metastore 服务。存储 Hive 表的元数据 (例如 schemas, locations 等)。
    *   **端口**: `9083` (用于 Hive Server 连接的 Thrift URI)
7.  `hive-server`:
    *   **镜像**: `bde2020/hive:2.3.2-postgresql-metastore`
    *   **用途**: HiveServer2。允许 JDBC/ODBC 客户端对 Hive 执行查询。
    *   **JDBC URL**: `jdbc:hive2://localhost:10000`
    *   **Web UI**: [http://localhost:10002](http://localhost:10002)

## 配置文件

-   `docker-compose.yml`: 定义所有服务、它们的镜像、端口、卷、依赖关系和环境变量。
-   `hadoop.env`: 包含主要用于 Hadoop 和 Hive 配置的环境变量，例如 HDFS 默认文件系统 URI 和 Hive Metastore URI。

## 如何运行

在您的终端中，导航到此目录 (`spark/docker/`)。

**启动服务:**

```bash
docker-compose up -d
```

**停止服务并移除容器、网络和卷:**

```bash
docker-compose down --volumes
```

**停止服务并移除容器和网络 (保留卷):**

```bash
docker-compose down
```

## 访问服务

-   **Hadoop HDFS NameNode UI**: [http://localhost:9870](http://localhost:9870)
-   **Spark Master UI**: [http://localhost:8080](http://localhost:8080)
-   **Spark Worker UI**: 通常可通过 Spark Master UI 上的链接访问 (例如, `http://localhost:8081`)。
-   **HiveServer2 Web UI**: [http://localhost:10002](http://localhost:10002)
-   **Hive JDBC 连接**: `jdbc:hive2://localhost:10000` (可用于 Beeline, DBeaver 等工具)

## 注意事项

-   **平台**: 所有服务都配置了 `platform: linux/amd64`。这确保了在 `arm64` 主机 (例如 Apple Silicon Macs) 上使用 Docker Desktop 的 Rosetta 2 仿真运行时具有兼容性。
-   **无 YARN**: 此配置不包含 Hadoop YARN。Hive 将以本地模式执行 MapReduce 作业，这适合学习，但对于较大的任务可能会较慢。Spark 以其独立集群模式运行。
-   **数据持久化**: 使用了命名卷 (`hadoop_namenode`, `hadoop_datanode`, `hive_metastore_db_data`) 来持久化 HDFS 数据和 Hive Metastore 数据，以便在容器重启后数据依然存在。如果运行 `docker-compose down --volumes`，这些数据将被移除。 