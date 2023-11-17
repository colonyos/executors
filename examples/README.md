# Deploy new code
colonies fs sync --dir src/ -l src

# Get the results
colonies fs sync --dir results/ -l results/ --keeplocal=false 
