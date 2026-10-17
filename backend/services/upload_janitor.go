package services

import (
	"log"
	"time"

	"ats-backend/config"
	"ats-backend/models"
)

const uploadClaimWindow = 24 * time.Hour

// StartUploadJanitor hourly removes files that were uploaded but never attached to an application,
// so abandoned CVs (personal data) don't accumulate in storage.
func StartUploadJanitor() {
	go func() {
		for {
			if err := purgeUnclaimedUploads(); err != nil {
				log.Printf("upload janitor: %v", err)
			}
			time.Sleep(time.Hour)
		}
	}()
}

func purgeUnclaimedUploads() error {
	var stale []models.Upload
	if err := config.DB.Where("claimed_at IS NULL AND created_at < ?", time.Now().Add(-uploadClaimWindow)).
		Limit(500).Find(&stale).Error; err != nil {
		return err
	}
	if len(stale) == 0 {
		return nil
	}
	byBucket := map[string][]string{}
	ids := make([]any, 0, len(stale))
	for _, up := range stale {
		byBucket[UploadBucket(up.Kind)] = append(byBucket[UploadBucket(up.Kind)], up.ObjectPath)
		ids = append(ids, up.ID)
	}
	for bucket, objects := range byBucket {
		if err := DeleteObjects(bucket, objects); err != nil {
			return err
		}
	}
	return config.DB.Where("id IN ?", ids).Delete(&models.Upload{}).Error
}

func UploadBucket(kind string) string {
	if kind == models.UploadPortfolio {
		return BucketPortfolios
	}
	return BucketResumes
}

// DeleteApplicationFiles removes an application's stored CV and portfolio; best effort.
func DeleteApplicationFiles(apps ...models.Application) {
	var resumes, portfolios []string
	for _, a := range apps {
		if a.ResumePath != nil {
			resumes = append(resumes, *a.ResumePath)
		}
		if a.PortfolioPath != nil {
			portfolios = append(portfolios, *a.PortfolioPath)
		}
	}
	for bucket, objects := range map[string][]string{BucketResumes: resumes, BucketPortfolios: portfolios} {
		if err := DeleteObjects(bucket, objects); err != nil {
			log.Printf("delete %d file(s) from %s: %v", len(objects), bucket, err)
			continue
		}
		if len(objects) > 0 {
			config.DB.Where("object_path IN ?", objects).Delete(&models.Upload{})
		}
	}
}
