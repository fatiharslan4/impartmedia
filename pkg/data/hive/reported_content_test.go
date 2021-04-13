// +build integration

package data

func (s *HiveTestSuite) TestAdminRetrieveReportedContent() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	postID := s.bootstrapPost(ctx, hiveID)
	//commentID := s.bootstrapComment(ctx, postID)

	reportedPosts, _, err := s.hiveData.GetUnreviewedReportedPosts(ctx, hiveID, 0)
	s.NoError(err)
	s.Len(reportedPosts, 0)

	reason := "because"
	err = s.hiveData.ReportPost(ctx, postID, &reason, false)
	s.NoError(err)

	reportedPosts, nextPage, err := s.hiveData.GetUnreviewedReportedPosts(ctx, hiveID, 0)
	s.NoError(err)
	s.Equal(1, nextPage.Offset)
	s.Len(reportedPosts, 1)

}
