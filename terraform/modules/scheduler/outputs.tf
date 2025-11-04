output "pubsub_topic_name" {
  description = "Pub/Subトピック名"
  value       = google_pubsub_topic.curator_trigger.name
}

output "pubsub_topic_id" {
  description = "Pub/SubトピックID（完全修飾名）"
  value       = google_pubsub_topic.curator_trigger.id
}

output "scheduler_job_name" {
  description = "Cloud Schedulerジョブ名"
  value       = google_cloud_scheduler_job.daily_curator.name
}
