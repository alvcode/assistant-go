package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"context"
)

type DriveFileChunkRepository interface {
	GetChunksSize(ctx context.Context, fileID int) (int64, error)
	Create(ctx context.Context, in *entity.DriveFileChunk) (*entity.DriveFileChunk, error)

	// includeRecycleBin - if true, then all files are returned;
	// if false, then files in the recycle bin will not be included in the result
	GetAllRecursive(
		ctx context.Context,
		structID int,
		userID int,
		includeRecycleBin bool,
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
	includeRecycleBin bool,
) ([]*entity.DriveFileChunk, error) {
	var query string
	if includeRecycleBin {
		query = `
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
	} else {
		query = `
			select * from drive_file_chunks dfc 
			where 
			dfc.drive_file_id in (
				select df.id from drive_files df 
				where 
				df.drive_struct_id in (
					WITH RECURSIVE structs AS (
						SELECT ds1.id
						FROM drive_structs ds1
						left join drive_recycle_bin drb1 on drb1.drive_struct_id = ds1.id
						WHERE drb1.id is null and ds1.id = $1 and ds1.user_id = $2
					
						UNION ALL
					
						SELECT ds2.id
						FROM drive_structs ds2
						left join drive_recycle_bin drb2 on drb2.drive_struct_id = ds2.id
						INNER JOIN structs s ON ds2.parent_id = s.id
						WHERE drb2.id is null
					)
					SELECT id FROM structs
				)
			)
		`
	}

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
