# Makefile for compiling gte.go

# Go compiler
GO := go

# Output executable name
TARGET := gte

# Source file
SOURCE := gte.go

all: $(TARGET)

$(TARGET): $(SOURCE)
	$(GO) build -o $(TARGET) $(SOURCE)

.PHONY: clean

clean:
	rm -f $(TARGET)
