job "cockroach" {
  namespace = "default"
  type = "system"
  region = "global"

  datacenters = [
    "home",
  ]

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "90s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    auto_promote = true
    canary = 1
    stagger = "30s"
  }

  group "databases" {
    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "http" {
        static = 3000
        to = 8080
      }
      port "tcp" {
        static = 26257
      }
    }

    task "serve" {
      driver = "exec"

      resources {
        cpu = 200
        memory = 256
      }

      service {
        name = "cockroach"

        port = "tcp"

        check {
          name = "alive"
          type = "tcp"
          port = "tcp"
          interval = "10s"
          timeout = "2s"

          check_restart {
            limit = 3
            grace = "60s"
            ignore_warnings = false
          }
        }
      }

      config {
        artifact {
          source = "https://binaries.cockroachdb.com/cockroach-v20.2.7.linux-amd64.tgz"
        }

        command = "cockroach"
        args = [
          "start",
          "--certs-dir=certs/",
          "--store=data",
          "--listen-addr=X.X.X.X:26257",
          "--join=X.X.X.X:26257,X.X.X.X:26258,X.X.X.X:26259"
        ]
      }

      logs {
        max_files = 10
        max_file_size = 2
      }
    }
  }
}
