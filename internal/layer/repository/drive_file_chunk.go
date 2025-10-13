package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"context"
)

type DriveFileChunkRepository interface {
	GetChunksSize(ctx context.Context, fileID int) (int64, error)
	Create(ctx context.Context, in *entity.DriveFileChunk) (*entity.DriveFileChunk, error)
	GetAllRecursive(
		ctx context.Context,
		structID int,
		userID int,
	) ([]*entity.DriveFileChunk, error)
	GetChunksInfo(ctx context.Context, fileID int) (*dto.DriveChunksInfo, error)
	GetByFileIDAndNumber(ctx context.Context, fileID int, chunkNumber int) (*entity.DriveFileChunk, error)
}

type driveFileChunkRepository struct {
	db DBExecutor
}

func NewDriveFileChunkRepository(db DBExecutor) DriveFileChunkRepository {
	return &driveFileChunkRepository{db: db}
}

func (r *driveFileChunkRepository) GetChunksSize(ctx context.Context, fileID int) (int64, error) {
	query := `SELECT 
    		coalesce(sum(dfc.size), 0) 
		FROM drive_file_chunks dfc
		WHERE dfc.drive_file_id = $1
	`

	var result int64
	err := r.db.QueryRow(ctx, query, fileID).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *driveFileChunkRepository) Create(ctx context.Context, in *entity.DriveFileChunk) (*entity.DriveFileChunk, error) {
	query := `
		INSERT INTO drive_file_chunks (drive_file_id, path, size, chunk_number) 
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	row := r.db.QueryRow(
		ctx,
		query,
		in.DriveFileID,
		in.Path,
		in.Size,
		in.ChunkNumber,
	)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (r *driveFileChunkRepository) GetAllRecursive(
	ctx context.Context,
	structID int,
	userID int,
) ([]*entity.DriveFileChunk, error) {
	query := `
		select * from drive_file_chunks dfc 
		where 
		dfc.drive_file_id in (
			select df.id from drive_files df 
			where 
			df.drive_struct_id in (
				WITH RECURSIVE structs AS (
					SELECT id
					FROM drive_structs 
					WHERE id = $1 and user_id = $2
				
					UNION ALL
				
					SELECT ds.id
					FROM drive_structs ds
					INNER JOIN structs s ON ds.parent_id = s.id
				)
				SELECT id FROM structs
			)
		)
	`

	rows, err := r.db.Query(ctx, query, structID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.DriveFileChunk, 0)

	for rows.Next() {
		dfc := &entity.DriveFileChunk{}
		if err := rows.Scan(&dfc.ID, &dfc.DriveFileID, &dfc.Path, &dfc.Size, &dfc.ChunkNumber); err != nil {
			return nil, err
		}
		result = append(result, dfc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *driveFileChunkRepository) GetChunksInfo(ctx context.Context, fileID int) (*dto.DriveChunksInfo, error) {
	query := `
		select  
			(select min(chunk_number) from drive_file_chunks dfc where drive_file_id = $1) as min_chunk_number,
			(select max(chunk_number) from drive_file_chunks dfc where drive_file_id = $1) as max_chunk_number
	`

	var result dto.DriveChunksInfo
	err := r.db.QueryRow(ctx, query, fileID).Scan(
		&result.StartNumber,
		&result.EndNumber,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *driveFileChunkRepository) GetByFileIDAndNumber(
	ctx context.Context,
	fileID int,
	chunkNumber int,
) (*entity.DriveFileChunk, error) {
	query := `select * from drive_file_chunks where drive_file_id = $1 and chunk_number = $2`

	var result entity.DriveFileChunk
	err := r.db.QueryRow(ctx, query, fileID, chunkNumber).Scan(
		&result.ID,
		&result.DriveFileID,
		&result.Path,
		&result.Size,
		&result.ChunkNumber,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
