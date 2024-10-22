package utils

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestSetZerologLevel(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    zerolog.Level
	}{
		{
			name:    "zerolog level trace",
			args:    args{"trace"},
			wantErr: false,
			want:    zerolog.TraceLevel,
		},
		{
			name:    "zerolog level debug",
			args:    args{"debug"},
			wantErr: false,
			want:    zerolog.DebugLevel,
		},
		{
			name:    "zerolog level info",
			args:    args{"info"},
			wantErr: false,
			want:    zerolog.InfoLevel,
		},
		{
			name:    "zerolog level warn",
			args:    args{"warn"},
			wantErr: false,
			want:    zerolog.WarnLevel,
		},
		{
			name:    "zerolog level error",
			args:    args{"error"},
			wantErr: false,
			want:    zerolog.ErrorLevel,
		},
		{
			name:    "zerolog level fatal",
			args:    args{"fatal"},
			wantErr: false,
			want:    zerolog.FatalLevel,
		},
		{
			name:    "zerolog level panic",
			args:    args{"panic"},
			wantErr: false,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level invalid",
			args:    args{"invalid"},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{""},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{" "},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{"\t"},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{"\n"},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{"\r"},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{"\r\n"},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
		{
			name:    "zerolog level empty",
			args:    args{"\r\n\t "},
			wantErr: true,
			want:    zerolog.PanicLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetZerologLevel(tt.args.level); (err != nil) != tt.wantErr {
				t.Errorf("SetZerologLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if zerolog.GlobalLevel() != tt.want {
				t.Errorf("SetZerologLevel() = %v, want %v", zerolog.GlobalLevel(), tt.want)
			}
		})
	}
}
