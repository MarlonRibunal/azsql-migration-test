-- Sample queries replayed against the Azure SQL Database Developer container.
-- Batches are separated by a standalone GO on its own line.

SELECT TOP 10 name, object_id FROM sys.tables ORDER BY name;
GO

SELECT COUNT(*) AS table_count FROM sys.tables;
GO
