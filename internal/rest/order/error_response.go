package order

type ErrorResponse struct {
	Code    string `json:"error_code"`
	Message string `json:"error_message"`
}

var (
	errRespInvalidPayload = ErrorResponse{
		Code:    "invalid_payload",
		Message: "Invalid payload.",
	}

	errRespOrderSize = ErrorResponse{
		Code:    "invalid_order_size",
		Message: "Order sizes must be greater than zero.",
	}

	errRespInvalidPackSizes = ErrorResponse{
		Code:    "invalid_pack_sizes",
		Message: "Pack sizes should have at least one size and all sizes should be greater than zero.",
	}

	errRespInternalServerError = ErrorResponse{
		Code:    "internal_server_error",
		Message: "Internal server error.",
	}
)
