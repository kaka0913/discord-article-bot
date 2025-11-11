output "scheduler_job_name" {
  description = "Cloud Schedulerジョブ名"
  value       = google_cloud_scheduler_job.daily_curator.name
}
