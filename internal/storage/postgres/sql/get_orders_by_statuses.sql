SELECT id, number, user_id, status, accrual, uploaded_at, processed_at
FROM orders
WHERE status = ANY($1)
ORDER BY uploaded_at ASC
