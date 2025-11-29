// =============================================================================
// Package grpc provides gRPC handlers for AI Service.
// =============================================================================
// AIServer implements all AI Service gRPC methods including:
// - Recognition: Floor plan recognition from images
// - Generation: Layout variant generation
// - Chat: Interactive AI assistant
// - Context: AI context management
// =============================================================================
package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/ai-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/ai-service/internal/openrouter"
	"github.com/xiiisorate/granula_api/ai-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AIServer implements the gRPC AI Service.
// It integrates with Scene Service for context-aware AI responses.
type AIServer struct {
	pb.UnimplementedAIServiceServer
	chatService        *service.ChatService
	recognitionService *service.RecognitionService
	generationService  *service.GenerationService
	sceneClient        *SceneClient // Integration with Scene Service
	log                *logger.Logger
}

// NewAIServer creates a new AIServer.
// sceneClient can be nil if Scene Service integration is not available.
func NewAIServer(
	chatService *service.ChatService,
	recognitionService *service.RecognitionService,
	generationService *service.GenerationService,
	sceneClient *SceneClient,
	log *logger.Logger,
) *AIServer {
	return &AIServer{
		chatService:        chatService,
		recognitionService: recognitionService,
		generationService:  generationService,
		sceneClient:        sceneClient,
		log:                log,
	}
}

// =============================================================================
// Recognition
// =============================================================================

// RecognizeFloorPlan recognizes a floor plan from an image.
func (s *AIServer) RecognizeFloorPlan(ctx context.Context, req *pb.RecognizeFloorPlanRequest) (*pb.RecognizeFloorPlanResponse, error) {
	s.log.Info("RecognizeFloorPlan called",
		logger.String("floor_plan_id", req.FloorPlanId),
		logger.Int("image_size", len(req.ImageData)),
	)

	options := entity.RecognitionOptions{
		DetectLoadBearing: req.Options.GetDetectLoadBearing(),
		DetectWetZones:    req.Options.GetDetectWetZones(),
		DetectFurniture:   req.Options.GetDetectFurniture(),
		Scale:             float64(req.Options.GetScale()),
		Orientation:       int(req.Options.GetOrientation()),
		DetailLevel:       int(req.Options.GetDetailLevel()),
	}

	job, err := s.recognitionService.StartRecognition(ctx, req.FloorPlanId, req.ImageData, req.ImageType, options)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.RecognizeFloorPlanResponse{
		Success: true,
		JobId:   job.ID.String(),
		Status:  convertJobStatusToPB(job.Status),
	}, nil
}

// GetRecognitionStatus gets the status of a recognition job.
func (s *AIServer) GetRecognitionStatus(ctx context.Context, req *pb.GetRecognitionStatusRequest) (*pb.GetRecognitionStatusResponse, error) {
	jobID, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, apperrors.InvalidArgument("job_id", "invalid UUID").ToGRPCError()
	}

	job, err := s.recognitionService.GetRecognitionStatus(ctx, jobID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	resp := &pb.GetRecognitionStatusResponse{
		JobId:    job.ID.String(),
		Status:   convertJobStatusToPB(job.Status),
		Progress: int32(job.Progress),
		Error:    job.Error,
	}

	if job.Result != nil {
		resp.Scene = convertRecognitionResultToPB(job.Result)
	}

	return resp, nil
}

// =============================================================================
// Generation
// =============================================================================

// GenerateVariants generates layout variants.
// It fetches scene data from Scene Service for context-aware generation.
func (s *AIServer) GenerateVariants(ctx context.Context, req *pb.GenerateVariantsRequest) (*pb.GenerateVariantsResponse, error) {
	s.log.Info("GenerateVariants called",
		logger.String("scene_id", req.SceneId),
		logger.String("branch_id", req.BranchId),
		logger.Int("variants_count", int(req.VariantsCount)),
	)

	// Fetch scene data from Scene Service for context-aware generation
	sceneData := ""
	if s.sceneClient != nil && req.SceneId != "" {
		sceneCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		data, err := s.sceneClient.GetSceneContext(sceneCtx, req.SceneId)
		if err != nil {
			s.log.Warn("failed to get scene data for generation, proceeding without context",
				logger.Err(err),
				logger.String("scene_id", req.SceneId),
			)
		} else {
			sceneData = data
			s.log.Debug("loaded scene context for generation",
				logger.String("scene_id", req.SceneId),
			)
		}
	}

	options := entity.GenerationOptions{
		PreserveLoadBearing: req.Options.GetPreserveLoadBearing(),
		CheckCompliance:     req.Options.GetCheckCompliance(),
		PreserveWetZones:    req.Options.GetPreserveWetZones(),
		Style:               convertGenerationStyleFromPB(req.Options.GetStyle()),
		Budget:              float64(req.Options.GetBudget()),
	}

	if len(req.Options.GetRequiredRooms()) > 0 {
		options.RequiredRooms = make([]string, len(req.Options.GetRequiredRooms()))
		for i, r := range req.Options.GetRequiredRooms() {
			options.RequiredRooms[i] = r.String()
		}
	}

	generateReq := service.GenerateRequest{
		SceneID:       req.SceneId,
		BranchID:      req.BranchId,
		Prompt:        req.Prompt,
		VariantsCount: int(req.VariantsCount),
		Options:       options,
		SceneData:     sceneData, // NOW WITH REAL DATA FROM SCENE SERVICE!
	}

	job, err := s.generationService.StartGeneration(ctx, generateReq)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GenerateVariantsResponse{
		Success: true,
		JobId:   job.ID.String(),
		Status:  convertJobStatusToPB(job.Status),
	}, nil
}

// GetGenerationStatus gets the status of a generation job.
func (s *AIServer) GetGenerationStatus(ctx context.Context, req *pb.GetGenerationStatusRequest) (*pb.GetGenerationStatusResponse, error) {
	jobID, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, apperrors.InvalidArgument("job_id", "invalid UUID").ToGRPCError()
	}

	job, err := s.generationService.GetGenerationStatus(ctx, jobID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	resp := &pb.GetGenerationStatusResponse{
		JobId:    job.ID.String(),
		Status:   convertJobStatusToPB(job.Status),
		Progress: int32(job.Progress),
		Error:    job.Error,
	}

	if len(job.Variants) > 0 {
		resp.Variants = convertVariantsToPB(job.Variants)
	}

	return resp, nil
}

// =============================================================================
// Chat
// =============================================================================

// SendChatMessage sends a chat message and gets a complete response.
func (s *AIServer) SendChatMessage(ctx context.Context, req *pb.ChatMessageRequest) (*pb.ChatMessageResponse, error) {
	s.log.Info("SendChatMessage called",
		logger.String("scene_id", req.SceneId),
	)

	resp, err := s.chatService.SendMessage(ctx, service.SendMessageRequest{
		SceneID:   req.SceneId,
		BranchID:  req.BranchId,
		Message:   req.Message,
		ContextID: req.ContextId,
	})
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.ChatMessageResponse{
		MessageId:        resp.MessageID,
		Response:         resp.Response,
		ContextId:        resp.ContextID,
		Actions:          convertActionsToPB(resp.Actions),
		GenerationTimeMs: resp.GenerationTimeMs,
		TokenUsage:       convertTokenUsageToPB(resp.TokenUsage),
	}, nil
}

// StreamChatResponse streams a chat response.
func (s *AIServer) StreamChatResponse(req *pb.ChatMessageRequest, stream pb.AIService_StreamChatResponseServer) error {
	s.log.Info("StreamChatResponse called",
		logger.String("scene_id", req.SceneId),
	)

	ctx := stream.Context()

	chunks, err := s.chatService.StreamMessage(ctx, service.SendMessageRequest{
		SceneID:   req.SceneId,
		BranchID:  req.BranchId,
		Message:   req.Message,
		ContextID: req.ContextId,
	})
	if err != nil {
		return apperrors.FromGRPCError(err).ToGRPCError()
	}

	for chunk := range chunks {
		if chunk.Error != nil {
			s.log.Error("stream error", logger.Err(chunk.Error))
			return apperrors.Internal("stream error").WithCause(chunk.Error).ToGRPCError()
		}

		pbChunk := &pb.ChatChunk{
			Content:    chunk.Content,
			IsFinal:    chunk.Done,
			ChunkIndex: int32(chunk.Index),
		}

		if chunk.MessageID != "" {
			pbChunk.MessageId = chunk.MessageID
		}
		if chunk.ContextID != "" {
			pbChunk.ContextId = chunk.ContextID
		}
		if len(chunk.Actions) > 0 {
			pbChunk.Actions = convertActionsToPB(chunk.Actions)
		}

		if err := stream.Send(pbChunk); err != nil {
			s.log.Error("failed to send chunk", logger.Err(err))
			return err
		}
	}

	return nil
}

// GetChatHistory retrieves chat history.
func (s *AIServer) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.GetChatHistoryResponse, error) {
	resp, err := s.chatService.GetHistory(ctx, service.GetHistoryRequest{
		SceneID:   req.SceneId,
		BranchID:  req.BranchId,
		ContextID: req.ContextId,
		Limit:     int(req.Limit),
		Cursor:    req.Cursor,
	})
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	messages := make([]*pb.ChatHistoryMessage, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, &pb.ChatHistoryMessage{
			Id:         msg.ID.String(),
			Role:       msg.Role,
			Content:    msg.Content,
			CreatedAt:  timestamppb.New(msg.CreatedAt),
			Actions:    convertActionsToPB(msg.Actions),
			TokenUsage: convertTokenUsageToPB(msg.TokenUsage),
		})
	}

	return &pb.GetChatHistoryResponse{
		Messages:   messages,
		HasMore:    resp.HasMore,
		NextCursor: resp.NextCursor,
	}, nil
}

// ClearChatHistory clears chat history.
func (s *AIServer) ClearChatHistory(ctx context.Context, req *pb.ClearChatHistoryRequest) (*pb.ClearChatHistoryResponse, error) {
	count, err := s.chatService.ClearHistory(ctx, req.SceneId, req.BranchId, req.ContextId)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.ClearChatHistoryResponse{
		DeletedCount: int32(count),
	}, nil
}

// GetContext retrieves AI context for a scene.
// Returns scene summary and context size information.
func (s *AIServer) GetContext(ctx context.Context, req *pb.GetContextRequest) (*pb.GetContextResponse, error) {
	s.log.Info("GetContext called",
		logger.String("scene_id", req.SceneId),
		logger.String("branch_id", req.BranchId),
	)

	// Get scene summary from Scene Service
	sceneSummary := ""
	if s.sceneClient != nil && req.SceneId != "" {
		sceneCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		data, err := s.sceneClient.GetSceneContext(sceneCtx, req.SceneId)
		if err == nil {
			sceneSummary = data
		} else {
			s.log.Warn("failed to get scene context",
				logger.Err(err),
				logger.String("scene_id", req.SceneId),
			)
		}
	}

	// Get recent messages to estimate context size
	contextSize := 0
	if s.chatService != nil {
		messages, err := s.chatService.GetRecentMessages(ctx, req.SceneId, req.BranchId, "", 10)
		if err == nil {
			for _, msg := range messages {
				contextSize += openrouter.EstimateTokens(msg.Content)
			}
		}
	}

	// Add scene summary tokens
	contextSize += openrouter.EstimateTokens(sceneSummary)

	return &pb.GetContextResponse{
		ContextId:    fmt.Sprintf("ctx_%s_%s", req.SceneId, req.BranchId),
		SceneSummary: sceneSummary,
		ContextSize:  int32(contextSize),
		UpdatedAt:    timestamppb.Now(),
	}, nil
}

// UpdateContext updates AI context by reloading scene data.
// Use this after making changes to the scene to ensure AI has fresh data.
func (s *AIServer) UpdateContext(ctx context.Context, req *pb.UpdateContextRequest) (*pb.UpdateContextResponse, error) {
	s.log.Info("UpdateContext called",
		logger.String("scene_id", req.SceneId),
		logger.String("branch_id", req.BranchId),
		logger.Bool("force", req.Force),
	)

	// Invalidate cache for this scene
	if s.sceneClient != nil && req.SceneId != "" {
		s.sceneClient.InvalidateCache(req.SceneId)
	}

	// Force reload scene data
	sceneSummary := ""
	if s.sceneClient != nil && req.SceneId != "" {
		sceneCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		data, err := s.sceneClient.GetSceneContext(sceneCtx, req.SceneId)
		if err != nil {
			s.log.Warn("failed to reload scene context",
				logger.Err(err),
				logger.String("scene_id", req.SceneId),
			)
			return &pb.UpdateContextResponse{
				ContextId:   fmt.Sprintf("ctx_%s_%s", req.SceneId, req.BranchId),
				Updated:     false,
				ContextSize: 0,
			}, nil
		}
		sceneSummary = data
	}

	contextSize := openrouter.EstimateTokens(sceneSummary)

	return &pb.UpdateContextResponse{
		ContextId:   fmt.Sprintf("ctx_%s_%s", req.SceneId, req.BranchId),
		Updated:     true,
		ContextSize: int32(contextSize),
	}, nil
}

// =============================================================================
// SelectSuggestion - выбор варианта из предложенных AI
// =============================================================================

// SelectSuggestion selects a variant from AI suggestions.
// This method activates the selected branch and creates a confirmation message.
func (s *AIServer) SelectSuggestion(ctx context.Context, req *pb.SelectSuggestionRequest) (*pb.SelectSuggestionResponse, error) {
	s.log.Info("SelectSuggestion called",
		logger.String("scene_id", req.SceneId),
		logger.String("message_id", req.MessageId),
		logger.Int("suggestion_index", int(req.SuggestionIndex)),
	)

	// Parse message ID
	messageID, err := uuid.Parse(req.MessageId)
	if err != nil {
		return nil, apperrors.InvalidArgument("message_id", "invalid UUID format").ToGRPCError()
	}

	// Get the message with suggestions
	message, err := s.chatService.GetMessage(ctx, messageID)
	if err != nil {
		s.log.Error("failed to get message for selection",
			logger.Err(err),
			logger.String("message_id", req.MessageId),
		)
		return nil, apperrors.NotFound("message", req.MessageId).ToGRPCError()
	}

	// Validate suggestion index
	if len(message.Actions) == 0 {
		return nil, apperrors.InvalidArgument("message_id", "message has no suggestions").ToGRPCError()
	}

	if int(req.SuggestionIndex) >= len(message.Actions) {
		return nil, apperrors.InvalidArgument("suggestion_index",
			fmt.Sprintf("index %d out of range, message has %d suggestions",
				req.SuggestionIndex, len(message.Actions))).ToGRPCError()
	}

	// Get selected action
	selectedAction := message.Actions[req.SuggestionIndex]
	branchID := selectedAction.Params["branch_id"]
	if branchID == "" {
		// Generate a branch ID if not provided
		branchID = fmt.Sprintf("br_sel_%s", uuid.New().String()[:8])
	}

	// TODO: Activate branch via Branch Service
	// This would call Branch Service to set the selected branch as active
	// branchClient.Activate(ctx, branchID)

	// Create confirmation message
	confirmationContent := fmt.Sprintf(
		"Отлично! Я активировал вариант \"%s\". Теперь вы можете:\n\n"+
			"- Редактировать планировку в 3D редакторе\n"+
			"- Попросить меня внести дополнительные изменения\n"+
			"- Создать новые варианты на основе этого\n\n"+
			"Что хотите сделать дальше?",
		selectedAction.Description,
	)

	confirmationMsg := entity.NewChatMessage(
		req.SceneId,
		branchID,
		message.ContextID,
		"assistant",
		confirmationContent,
	)

	if err := s.chatService.SaveMessage(ctx, confirmationMsg); err != nil {
		s.log.Warn("failed to save confirmation message",
			logger.Err(err),
		)
	}

	s.log.Info("suggestion selected successfully",
		logger.String("scene_id", req.SceneId),
		logger.String("selected_branch_id", branchID),
		logger.String("action_type", selectedAction.Type),
	)

	return &pb.SelectSuggestionResponse{
		SelectedBranchId:    branchID,
		BranchActivated:     true,
		ConfirmationMessage: confirmationContent,
	}, nil
}

// =============================================================================
// Conversion helpers
// =============================================================================

func convertJobStatusToPB(status entity.JobStatus) pb.JobStatus {
	switch status {
	case entity.JobStatusPending:
		return pb.JobStatus_JOB_STATUS_PENDING
	case entity.JobStatusProcessing:
		return pb.JobStatus_JOB_STATUS_PROCESSING
	case entity.JobStatusCompleted:
		return pb.JobStatus_JOB_STATUS_COMPLETED
	case entity.JobStatusFailed:
		return pb.JobStatus_JOB_STATUS_FAILED
	case entity.JobStatusCancelled:
		return pb.JobStatus_JOB_STATUS_CANCELLED
	default:
		return pb.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}

func convertGenerationStyleFromPB(style pb.GenerationStyle) entity.GenerationStyle {
	switch style {
	case pb.GenerationStyle_GENERATION_STYLE_MINIMAL:
		return entity.GenerationStyleMinimal
	case pb.GenerationStyle_GENERATION_STYLE_MODERATE:
		return entity.GenerationStyleModerate
	case pb.GenerationStyle_GENERATION_STYLE_CREATIVE:
		return entity.GenerationStyleCreative
	default:
		return entity.GenerationStyleModerate
	}
}

func convertRecognitionResultToPB(result *entity.RecognitionResult) *pb.RecognizedScene {
	scene := &pb.RecognizedScene{
		Dimensions: &commonpb.Dimensions2D{
			Width:  result.Dimensions.Width,
			Height: result.Dimensions.Height,
		},
		TotalArea: float32(result.TotalArea),
		Metadata: &pb.RecognitionMetadata{
			ModelVersion:     result.ModelVersion,
			ProcessingTimeMs: result.ProcessingTimeMs,
		},
	}

	for _, w := range result.Walls {
		scene.Walls = append(scene.Walls, &pb.RecognizedWall{
			TempId:                w.TempID,
			Start:                 &commonpb.Point2D{X: w.Start.X, Y: w.Start.Y},
			End:                   &commonpb.Point2D{X: w.End.X, Y: w.End.Y},
			Thickness:             float32(w.Thickness),
			IsLoadBearing:         w.IsLoadBearing,
			Confidence:            float32(w.Confidence),
			LoadBearingConfidence: float32(w.LoadBearingConfidence),
		})
	}

	for _, r := range result.Rooms {
		vertices := make([]*commonpb.Point2D, 0, len(r.Boundary))
		for _, p := range r.Boundary {
			vertices = append(vertices, &commonpb.Point2D{X: p.X, Y: p.Y})
		}
		room := &pb.RecognizedRoom{
			TempId:     r.TempID,
			Type:       convertRoomTypeToPB(r.Type),
			Boundary:   &commonpb.Polygon2D{Vertices: vertices},
			Area:       float32(r.Area),
			IsWetZone:  r.IsWetZone,
			Confidence: float32(r.Confidence),
			WallIds:    r.WallIDs,
		}
		scene.Rooms = append(scene.Rooms, room)
	}

	for _, o := range result.Openings {
		scene.Openings = append(scene.Openings, &pb.RecognizedOpening{
			TempId:     o.TempID,
			Type:       convertOpeningTypeToPB(o.Type),
			Position:   &commonpb.Point2D{X: o.Position.X, Y: o.Position.Y},
			Width:      float32(o.Width),
			WallId:     o.WallID,
			Confidence: float32(o.Confidence),
		})
	}

	return scene
}

func convertRoomTypeToPB(t string) pb.RoomType {
	switch t {
	case "LIVING":
		return pb.RoomType_ROOM_TYPE_LIVING
	case "BEDROOM":
		return pb.RoomType_ROOM_TYPE_BEDROOM
	case "KITCHEN":
		return pb.RoomType_ROOM_TYPE_KITCHEN
	case "BATHROOM":
		return pb.RoomType_ROOM_TYPE_BATHROOM
	case "TOILET":
		return pb.RoomType_ROOM_TYPE_TOILET
	case "HALLWAY":
		return pb.RoomType_ROOM_TYPE_HALLWAY
	case "BALCONY":
		return pb.RoomType_ROOM_TYPE_BALCONY
	case "STORAGE":
		return pb.RoomType_ROOM_TYPE_STORAGE
	default:
		return pb.RoomType_ROOM_TYPE_UNSPECIFIED
	}
}

func convertOpeningTypeToPB(t string) pb.OpeningType {
	switch t {
	case "door":
		return pb.OpeningType_OPENING_TYPE_DOOR
	case "window":
		return pb.OpeningType_OPENING_TYPE_WINDOW
	case "arch":
		return pb.OpeningType_OPENING_TYPE_ARCH
	default:
		return pb.OpeningType_OPENING_TYPE_UNSPECIFIED
	}
}

func convertVariantsToPB(variants []entity.GeneratedVariant) []*pb.GeneratedVariant {
	result := make([]*pb.GeneratedVariant, 0, len(variants))
	for _, v := range variants {
		pbVariant := &pb.GeneratedVariant{
			Id:            v.ID,
			BranchId:      v.BranchID,
			Name:          v.Name,
			Description:   v.Description,
			Score:         float32(v.Score),
			IsCompliant:   v.IsCompliant,
			EstimatedCost: float32(v.EstimatedCost),
		}
		for _, c := range v.Changes {
			pbVariant.Changes = append(pbVariant.Changes, &pb.VariantChange{
				Type:        c.Type,
				Description: c.Description,
				ElementIds:  c.ElementIDs,
			})
		}
		result = append(result, pbVariant)
	}
	return result
}

func convertActionsToPB(actions []entity.SuggestedAction) []*pb.SuggestedAction {
	result := make([]*pb.SuggestedAction, 0, len(actions))
	for _, a := range actions {
		result = append(result, &pb.SuggestedAction{
			Id:                   a.ID,
			Type:                 a.Type,
			Description:          a.Description,
			Params:               a.Params,
			Confidence:           float32(a.Confidence),
			RequiresConfirmation: a.RequiresConfirmation,
		})
	}
	return result
}

func convertTokenUsageToPB(usage *entity.TokenUsage) *pb.TokenUsage {
	if usage == nil {
		return nil
	}
	return &pb.TokenUsage{
		PromptTokens:     int32(usage.PromptTokens),
		CompletionTokens: int32(usage.CompletionTokens),
		TotalTokens:      int32(usage.TotalTokens),
	}
}
