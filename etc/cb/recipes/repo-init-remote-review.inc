# Include for initializing the remove reviews.
# Must be in ${dir}
step = info "setting up git review remote for the cryptomanager project"
step = exec git --no-pager remote add review ${revurl}/cryptomanager

step = info "setting up git review remote for the cmservices project"
step = cd devices/cmservice
step = exec git --no-pager remote add review ${revurl}/cmservices
step = cd ../..

step = info "setting up git review remote for the tools project"
step = cd tools
step = exec git --no-pager remote add review ${revurl}/tools
step = cd ..

step = info "setting up git review remote for the vmware-automation project"
step = cd vmware-automation
step = exec git --no-pager remote add review ${revurl}/vmware-automation
step = cd ..
