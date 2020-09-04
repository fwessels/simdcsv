#define CREATE_MASK(_Y1, _Y2, _R1, _R2) \
	VPMOVMSKB _Y1, _R1 \
	VPMOVMSKB _Y2, _R2 \
	SHLQ      $32, _R2 \
	ORQ       _R1, _R2
