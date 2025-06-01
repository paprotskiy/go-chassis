ci_cd_page="https://ci-cd.page"
envs=`grep -v "*" .env`
git_curr_branch_name=`git rev-parse --abbrev-ref HEAD`
semver_tag_latest=`git tag -l 'v*' | sed 's/^v//' | sort -t '.' -k1,1n -k2,2n -k3,3n | sed 's/^/v/' | tail -n 1`
semver_tag_inc_patch=`echo $$latest_tag | sed 's/\([0-9]\+\.[0-9]\+\)\.[0-9]\+/\1.$$latest_patch/'`


safe-run:
	VERSION=$(semver_tag_latest) docker-compose up --build go-chassis

watch:
	go run src/cmd/testwatcher/test-watcher.go

local-run:
	export $(envs) && go run src/cmd/app/main.go

retag-latest:
	branch=$(git_curr_branch_name) &&\
	(:&&\
#		(if echo $$branch | grep -q "^task-.*"; then exit 0; else echo; echo; echo "!! only branches, based on tasks can be retaged"; exit 127; fi) && \
		(tagname="$$branch-latest";\
			(echo ">> delete tag locally"         && git tag -d $$tagname);\
			(echo ">> delete tag remotely"        && git push --delete origin $$tagname);\
			(echo ">> create tag locally"         && git tag $$tagname);\
			(echo ">>>> push tag to remote"       && git push origin $$tagname );\
			(echo ">>>> open ci/cd piplanes page" && xdg-open $(ci_cd_page))\
		)\
	)
	
# ensure that at least the v0.0.0 tag exists
patch:
	@latest_tag=$(semver_tag_latest) &&\
	(:&&\
		(patch_new_ver=$$(($$(echo $$latest_tag | awk -F. '{print $$NF}') + 1)) ;\
			regex='s/\([0-9]*\.[0-9]*\)\.[0-9]*$$/\1.'$$patch_new_ver'/' ;\
			new_tag=$$(echo $$latest_tag | sed "$$regex") ;\
			(echo ">> create tag locally: "$$new_tag && git tag $$new_tag)\
		) ;\
	)

# to install golang migrate as CLI
# go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
new-migration:
	@while [ -z "$$NAME" ]; do \
		read -p "Enter migration name: " NAME; \
	done; \
	migrate create -ext sql -dir src/internal/adapters/storage/migrations $$NAME
	exit 0

# go install github.com/ofabry/go-callvis@latest
call-graph:
	$(HOME)/go/bin/go-callvis -group pkg ./...
	
# go install github.com/kisielk/godepgraph@latest
deps-graph:
	$(HOME)/go/bin/godepgraph -novendor -s -p github.com ./src/cmd/app/ > deps.dot
