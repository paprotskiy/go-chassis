watch:
	TESTCONTAINERS_RYUK_DISABLED=true \
	$(HOME)/go/bin/reflex -r '\.go$$' -- sh -c 'go test ./... && echo -e "\033[32m\n\n ---- OK ---- \n\033[0m" || echo -e "\033[31m\n\n ---- FAIL ---- \n\033[0m"'
