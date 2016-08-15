# Initialize the repo.
step = must-not-exist-dir ${dir}
step = exec mkdir -m 0777 ${dir}

step = info "initializing the repo"
step = cd ${dir}
step = exec repo init -u ${manurl} -m ${branch}.xml
step = exec repo sync

