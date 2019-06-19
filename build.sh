echo 'Downloading frontend ...'
curl -L -O https://github.com/alxarno/swap-ui/releases/download/0.0.1/dist.tar.gz --silent
echo 'Unpacking frontend ...'
tar xzf dist.tar.gz
mv dist ui
packr
mkdir releases
case "$(uname -s)" in

   Darwin)
     echo 'Mac OS X'
     go build -o ./releases/swap-darwin
     ;;

   Linux)
    echo 'Building for Linux based OS ...'
    go build -o ./releases/swap-linux
     ;;

   CYGWIN*|MINGW32*|MSYS*|MINGW64*)
      echo 'Building for Windows ...'
      go build -o ./releases/swap.exe
     ;;

   *)
     echo 'other OS' 
      go build -o ./releases/swap-some-os
     ;;
esac
echo 'Clearing ...'
packr clean
rm -rf ui
rm -rf dist
rm -rf dist.tar.gz