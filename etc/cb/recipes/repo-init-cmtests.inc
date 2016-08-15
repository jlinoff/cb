# Must be in ${dir}.
step = info "initializing cmtests"
step = cd cmtests
step = exec make
step = exec scripts/personalize.sh
