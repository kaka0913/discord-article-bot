output "database_name" {
  description = "Firestoreデータベース名"
  value       = google_firestore_database.database.name
}

output "database_id" {
  description = "FirestoreデータベースID"
  value       = google_firestore_database.database.id
}
