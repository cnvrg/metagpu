package metagpusrv

import (
	"context"
	"errors"
	"fmt"
	devicevpb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
	devicevapi "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/metagpusrv/deviceapi/device/v1"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"path/filepath"
	"time"
)

type VisibilityLevel string

type MetaGpuServer struct {
	plugin                        *deviceplugin.MetaGpuDevicePlugin
	ContainerLevelVisibilityToken string
	DeviceLevelVisibilityToken    string
}

var (
	ContainerVisibility      VisibilityLevel = "containerVisibilityLevelToken"
	DeviceVisibility         VisibilityLevel = "deviceVisibilityLevelToken"
	TokenVisibilityClaimName                 = "visibilityLevel"
)

func NewMetaGpuServer(plugin *deviceplugin.MetaGpuDevicePlugin) *MetaGpuServer {
	s := &MetaGpuServer{plugin: plugin}
	s.ContainerLevelVisibilityToken = s.GenerateAuthTokens(ContainerVisibility)
	s.DeviceLevelVisibilityToken = s.GenerateAuthTokens(DeviceVisibility)
	plugin.SetContainerLevelVisibilityToken(s.ContainerLevelVisibilityToken)
	plugin.SetDeviceLevelVisibilityToken(s.DeviceLevelVisibilityToken)
	s.SaveTokensOnLocalStorage()
	return s
}

func (s *MetaGpuServer) Start() {

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s", viper.GetString("metagpu-server-addr")))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Infof("grpc server listening on %s", viper.GetString("metagpu-server-addr"))

		opts := []grpc.ServerOption{grpc.UnaryInterceptor(s.unaryServerInterceptor())}

		grpcServer := grpc.NewServer(opts...)

		dp := devicevapi.DeviceService{}
		devicevpb.RegisterDeviceServiceServer(grpcServer, &dp)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *MetaGpuServer) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		visibility, err := authorize(ctx)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, TokenVisibilityClaimName, visibility)
		ctx = context.WithValue(ctx, "containerVl", string(ContainerVisibility))
		ctx = context.WithValue(ctx, "deviceVl", string(DeviceVisibility))
		ctx = context.WithValue(ctx, "plugin", s.plugin)
		h, err := handler(ctx, req)
		log.Infof("[method: %s duration: %s]", info.FullMethod, time.Since(start))
		return h, err
	}
}

func (s *MetaGpuServer) GenerateAuthTokens(visibility VisibilityLevel) string {

	claims := jwt.MapClaims{"email": "metagpu@instance", TokenVisibilityClaimName: visibility}
	containerScopeToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := containerScopeToken.SignedString([]byte(viper.GetString("jwtSecret")))
	if err != nil {
		log.Error(err)
	}
	return tokenString
}

func (s *MetaGpuServer) SaveTokensOnLocalStorage() {
	ex, err := os.Executable()
	if err != nil {
		log.Error(err)
		return
	}
	exPath := filepath.Dir(ex)
	filePath := exPath + "/.mgsrvtokens"
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		log.Error("unable write tokens to local storage")
	}
	tokensStr := fmt.Sprintf("containerVl: %s\ndeviceVl:%s\n", s.ContainerLevelVisibilityToken, s.DeviceLevelVisibilityToken)
	_, err = f.WriteString(tokensStr)
	if err != nil {
		log.Error("failed to write tokens to .mgsrvtokens file")
	}
	log.Infof("tokens saved in %s", filePath)

}

func authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "retrieving metadata is failed")
	}

	authHeader, ok := md["authorization"]

	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not supplied")
	}

	tokenString := authHeader[0]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())
		}
		return []byte(viper.GetString("jwtSecret")), nil
	})
	if err != nil {
		log.Error(err)
		return "", status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if visibility, ok := claims[TokenVisibilityClaimName]; ok {
			visibility := visibility.(string)
			return visibility, nil
		}
	}
	return "", status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())

}
