param (
    $stage = 'build'
)

$version = git tag -l --sort=-v:refname | Select-Object -First 1
$build_at = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"
$commit = git log --max-count=1 --pretty=format:%aI_%h

$ldflags = "-w -s -X main.version=$version -X main.build=$build_at -X main.commit=$commit"

$app_name = 'subs-api-v6.exe'
$out_dir = './build'

$exec = "$out_dir/$app_name"

switch ($stage)
{
    'build' { go build -o $exec -ldflags $ldflags -tags production -v . }
    'run' { "$exec -production=false -livemode=false" | Invoke-Expression }
}
