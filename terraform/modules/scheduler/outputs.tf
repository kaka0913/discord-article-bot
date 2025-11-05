output "pubsub_topic_id" {
  description = "Pub/SubトピックID（完全修飾名） - Cloud Functionsのトリガーに使用"
  value       = google_pubsub_topic.curator_trigger.id
}
