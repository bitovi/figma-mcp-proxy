resource "aws_cloudwatch_dashboard" "figma_mcp_proxy" {
  dashboard_name = "${local.fully_qualified_name}"

  dashboard_body = jsonencode({
    widgets = [
      # Top row - Current values (percentages)
      {
        type = "metric"
        x    = 0
        y    = 0
        width = 6
        height = 3
        properties = {
          metrics = [
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryUtilization\" objectname=\"Memory\"', 'Average', 60)", id = "m1" }]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Committed Memory (%)"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type = "metric"
        x    = 6
        y    = 0
        width = 6
        height = 3
        properties = {
          metrics = [
            [{ expression = "(1 - (m1 / 7974)) * 100", label = "Physical RAM Used %", id = "e1" }],
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryAvailableMBytes\" objectname=\"Memory\"', 'Average', 60)", id = "m1", visible = false }]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Physical RAM Used (%)"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type = "metric"
        x    = 12
        y    = 0
        width = 6
        height = 3
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", "InstanceId", aws_instance.win2025.id]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Current CPU (%)"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type = "metric"
        x    = 18
        y    = 0
        width = 6
        height = 3
        properties = {
          metrics = [
            [{ expression = "100 - m1", label = "Disk Used %", id = "e1" }],
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,instance,objectname} MetricName=\"DiskFreeSpace\" objectname=\"LogicalDisk\" instance=\"C:\"', 'Average', 60)", id = "m1", visible = false }]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Current Disk (%)"
          period  = 60
          stat    = "Average"
        }
      },
      # Second row - Actual units (MB)
      {
        type = "metric"
        x    = 0
        y    = 3
        width = 6
        height = 3
        properties = {
          metrics = [
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryAvailableMBytes\" objectname=\"Memory\"', 'Average', 60)", id = "m1" }]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Available RAM (MB)"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type = "text"
        x    = 6
        y    = 3
        width = 6
        height = 3
        properties = {
          markdown = "## 7974 MB\nTotal RAM"
        }
      },
      {
        type = "text"
        x    = 12
        y    = 3
        width = 6
        height = 3
        properties = {
          markdown = "## 2 vCPUs\nt3.large"
        }
      },
      {
        type = "metric"
        x    = 18
        y    = 3
        width = 6
        height = 3
        properties = {
          metrics = [
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,instance,objectname} MetricName=\"DiskFreeMBytes\" objectname=\"LogicalDisk\" instance=\"C:\"', 'Average', 60)", id = "m1" }]
          ]
          view    = "singleValue"
          region  = "us-east-2"
          title   = "Free Disk Space (MB)"
          period  = 60
          stat    = "Average"
        }
      },
      # Third Row - Target Health (full width)
      {
        type = "metric"
        x    = 0
        y    = 6
        width = 24
        height = 6
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "HealthyHostCount", "TargetGroup", aws_lb_target_group.app.arn_suffix, "LoadBalancer", aws_lb.app.arn_suffix, { stat = "Average" }],
            [".", "UnHealthyHostCount", ".", ".", ".", ".", { stat = "Average" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Target Health"
          period  = 60
          yAxis = {
            left = {
              min = 0
              label = "Host Count"
            }
          }
        }
      },
      # Fourth Row - Memory Utilization (full width)
      {
        type = "metric"
        x    = 0
        y    = 12
        width = 24
        height = 6
        properties = {
          metrics = [
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryUtilization\" objectname=\"Memory\"', 'Average', 60)", id = "m1" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Memory Utilization (% of committed bytes in use, RAM + page file)"
          period  = 60
          stat    = "Average"
          yAxis = {
            left = {
              min = 0
              max = 100
              label = "Percent"
            }
          }
        }
      },
      # Fifth Row - CPU Utilization (full width)
      {
        type = "metric"
        x    = 0
        y    = 18
        width = 24
        height = 6
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", "InstanceId", aws_instance.win2025.id]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "EC2 CPU Utilization"
          period  = 60
          stat    = "Average"
          yAxis = {
            left = {
              min = 0
              max = 100
              label = "Percent"
            }
          }
        }
      },
      # Sixth Row Left - Available RAM
      {
        type = "metric"
        x    = 0
        y    = 24
        width = 12
        height = 6
        properties = {
          metrics = [
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryAvailableMBytes\" objectname=\"Memory\"', 'Average', 60)", id = "m1" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Available Physical RAM (MB)"
          period  = 60
          stat    = "Average"
          yAxis = {
            left = {
              min = 0
              label = "Megabytes"
            }
          }
        }
      },
      # Sixth Row Right - Physical RAM Utilization
      {
        type = "metric"
        x    = 12
        y    = 24
        width = 12
        height = 6
        properties = {
          metrics = [
            [{ expression = "(1 - (m1 / 7974)) * 100", label = "Physical RAM Used %", id = "e1" }],
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,objectname} MetricName=\"MemoryAvailableMBytes\" objectname=\"Memory\"', 'Average', 60)", id = "m1", visible = false }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Physical RAM Utilization (% of 8GB total)"
          period  = 60
          stat    = "Average"
          yAxis = {
            left = {
              min = 0
              max = 100
              label = "Percent"
            }
          }
        }
      },
      # Seventh Row Left - Load Balancer Metrics
      {
        type = "metric"
        x    = 0
        y    = 30
        width = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", aws_lb.app.arn_suffix, { stat = "Average" }],
            [".", "RequestCount", ".", ".", { stat = "Sum", yAxis = "right" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Load Balancer Metrics"
          period  = 60
          yAxis = {
            left = {
              label = "Response Time (seconds)"
            }
            right = {
              label = "Request Count"
            }
          }
        }
      },
      # Seventh Row Right - Disk Usage
      {
        type = "metric"
        x    = 12
        y    = 30
        width = 12
        height = 6
        properties = {
          metrics = [
            [{ expression = "100 - m1", label = "Disk Used %", id = "e1" }],
            [{ expression = "SEARCH('{CWAgent/${var.target_environment},host,instance,objectname} MetricName=\"DiskFreeSpace\" objectname=\"LogicalDisk\" instance=\"C:\"', 'Average', 60)", id = "m1", visible = false }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = "us-east-2"
          title   = "Disk Utilization - C: Drive (%)"
          period  = 60
          stat    = "Average"
          yAxis = {
            left = {
              min = 0
              max = 100
              label = "Percent"
            }
          }
        }
      },
      # Eighth Row - Application Logs
      {
        type = "log"
        x    = 0
        y    = 36
        width = 24
        height = 6
        properties = {
          query   = "SOURCE '/aws/ec2/figma-proxy-${var.target_environment}'\n| fields @timestamp, @message\n| sort @timestamp desc\n| limit 100"
          region  = "us-east-2"
          title   = "FigmaProxy Application Logs (Last 100)"
        }
      }
    ]
  })
}
