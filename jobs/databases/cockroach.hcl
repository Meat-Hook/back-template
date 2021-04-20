job "cockroach" {
  namespace = "default"
  type = "service"
  region = "global"

  datacenters = [
    "home",
  ]

  constraint {
    attribute = "${attr.unique.hostname}"
    value = "home-server"
  }

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "90s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    stagger = "30s"
  }

  group "master" {
    count = 1

    volume "certs" {
      type = "host"
      read_only = true
      source = "cockroach-certs"
    }

    volume "data" {
      type = "host"
      read_only = false
      source = "cockroach-data"
    }

    ephemeral_disk {
      migrate = true
      size = 10240
      sticky = true
    }

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "dashboard" {
        static = 8080
      }
      port "tcp" {
        static = 26257
      }
      dns {
        servers = [
          "192.168.31.116",
        ]
      }
    }

    // admin dashboard
    service {
      name = "cockroach-dashboard"

      port = "dashboard"

      tags = [
        "database",
        "docker",
        "admin",
      ]

      check {
        name = "alive-dashboard"
        type = "tcp"
        port = "dashboard"
        interval = "10s"
        timeout = "2s"

        check_restart {
          limit = 3
          grace = "60s"
          ignore_warnings = false
        }
      }
    }

    // database
    service {
      name = "cockroach-master"

      port = "tcp"

      tags = [
        "database",
        "master",
        "docker",
      ]

      check {
        name = "alive-tcp"
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

    task "serve" {
      driver = "docker"

      resources {
        cpu = 1000
        memory = 1024
      }

      config {
        image = "cockroachdb/cockroach:v20.2.7"

        hostname = "cockroach-master.service.consul"

        ports = [
          "dashboard",
          "tcp",
        ]

        args = [
          "start",
          "--certs-dir",
          "/opt/cockroach/certs/master",
          "--store",
          "/opt/cockroach/data/master/",
          "--host",
          "cockroach-master.service.consul",
          "--port",
          "${NOMAD_PORT_tcp}",
          "--http-port",
          "${NOMAD_PORT_dashboard}",
          "--join",
          "${COCKROACH_ADDRS}"
        ]
      }

      volume_mount {
        volume = "certs"
        destination = "/opt/cockroach/certs"
        read_only = true
      }

      volume_mount {
        volume = "data"
        destination = "/opt/cockroach/data/"
        read_only = false
      }

      logs {
        max_files = 10
        max_file_size = 2
      }

      template {
        data = <<EOH
COCKROACH_ADDRS = cockroach-master.service.consul:26257{{ range service "cockroach-slave" }},{{ .Address }}:{{ .Port }}{{ end }}
EOH

        destination = "local/config.env"
        change_mode = "restart"
        env         = true
      }
    }
  }

  group "slave" {
    count = 2

    volume "certs" {
      type = "host"
      read_only = true
      source = "cockroach-certs"
    }

    volume "data" {
      type = "host"
      read_only = false
      source = "cockroach-data"
    }

    ephemeral_disk {
      migrate = true
      size = 10240
      sticky = true
    }

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "dashboard" {
        to = 8080
      }
      port "tcp" {
        to = 26257
      }
      dns {
        servers = [
          "192.168.31.116",
        ]
      }
    }

    // database
    service {
      name = "cockroach-slave"

      port = "tcp"

      tags = [
        "database",
        "slave",
        "docker",
      ]

      check {
        name = "alive-tcp"
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

    task "serve" {
      driver = "docker"

      resources {
        cpu = 500
        memory = 512
      }

      config {
        image = "cockroachdb/cockroach:v20.2.7"

        hostname = "cockroach-slave.service.consul"

        ports = [
          "dashboard",
          "tcp",
        ]

        args = [
          "start",
          "--certs-dir",
          "/opt/cockroach/certs/slave",
          "--store",
          "/opt/cockroach/data/node-${NOMAD_ALLOC_INDEX}",
          "--host",
          "cockroach-slave.service.consul",
          "--port",
          "${NOMAD_PORT_tcp}",
          "--http-port",
          "${NOMAD_PORT_dashboard}",
          "--join",
          "${COCKROACH_ADDRS}"
        ]
      }

      volume_mount {
        volume = "certs"
        destination = "/opt/cockroach/certs"
        read_only = true
      }

      volume_mount {
        volume = "data"
        destination = "/opt/cockroach/data/"
        read_only = false
      }

      logs {
        max_files = 10
        max_file_size = 2
      }

      template {
        data = <<EOH
COCKROACH_ADDRS = cockroach-master.service.consul:26257{{ range service "cockroach-slave" }},cockroach-slave.service.consul:{{ .Port }}{{ end }}
EOH

        destination = "local/config.env"
        change_mode = "restart"
        env         = true
      }
    }
  }
}
