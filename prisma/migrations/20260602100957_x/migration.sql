-- CreateEnum
CREATE TYPE "deployment_status" AS ENUM ('IDLE', 'IN_PROGRESS', 'SUCCESSFUL', 'FAILED');

-- CreateEnum
CREATE TYPE "machine_status" AS ENUM ('CREATING', 'RUNNING', 'REBOOTING', 'SHUTTING_DOWN', 'SHUTDOWN');

-- CreateTable
CREATE TABLE "users" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "first_name" TEXT,
    "last_name" TEXT,
    "email" TEXT,
    "password" TEXT,
    "roles" TEXT[],

    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "deployment_logs" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "deployment_id" TEXT NOT NULL,
    "job_id" TEXT,
    "message" TEXT,

    CONSTRAINT "deployment_logs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "deployments" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "machine_id" TEXT NOT NULL,
    "repo_connection_id" TEXT NOT NULL,
    "commit_hash" TEXT,
    "commit_message" TEXT,
    "actor" TEXT,
    "status" "deployment_status" NOT NULL DEFAULT 'IDLE',

    CONSTRAINT "deployments_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "github_account_connections" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "user_id" TEXT NOT NULL,
    "github_id" TEXT,
    "github_installation_id" INTEGER,
    "github_username" TEXT,
    "github_email" TEXT,

    CONSTRAINT "github_account_connections_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "machines" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "owner_id" TEXT NOT NULL,
    "container_id" TEXT,
    "machine_name" TEXT,
    "machine_image" TEXT,
    "cpu" TEXT,
    "memory" TEXT,
    "machine_status" "machine_status" NOT NULL DEFAULT 'CREATING',

    CONSTRAINT "machines_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "networks" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "machine_id" TEXT NOT NULL,
    "host_name" TEXT,
    "protocol" TEXT,
    "listening_port" INTEGER NOT NULL,
    "destination_port" INTEGER NOT NULL,
    "service_id" TEXT,
    "ingress_id" TEXT,

    CONSTRAINT "networks_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "repo_connections" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "deleted_at" TIMESTAMP(3) DEFAULT null,
    "machine_id" TEXT NOT NULL,
    "repo_id" TEXT,
    "repoName" TEXT,

    CONSTRAINT "repo_connections_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "users_email_key" ON "users"("email");

-- CreateIndex
CREATE UNIQUE INDEX "machines_owner_id_machine_name_key" ON "machines"("owner_id", "machine_name");

-- AddForeignKey
ALTER TABLE "deployment_logs" ADD CONSTRAINT "deployment_logs_deployment_id_fkey" FOREIGN KEY ("deployment_id") REFERENCES "deployments"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "deployments" ADD CONSTRAINT "deployments_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machines"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "deployments" ADD CONSTRAINT "deployments_repo_connection_id_fkey" FOREIGN KEY ("repo_connection_id") REFERENCES "repo_connections"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "github_account_connections" ADD CONSTRAINT "github_account_connections_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "machines" ADD CONSTRAINT "machines_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "networks" ADD CONSTRAINT "networks_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machines"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "repo_connections" ADD CONSTRAINT "repo_connections_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machines"("id") ON DELETE CASCADE ON UPDATE CASCADE;
