curl -L -O https://github.com/alxarno/swap-ui/releases/download/0.0.1/dist.tar.gz --silent
tar xzf dist.tar.gz
mv dist ui
rm -rf dist
rm -rf dist.tar.gz
packr
mkdir releases
case "$(uname -s)" in

   Darwin)
     echo 'Mac OS X'
     ;;

   Linux)
    echo 'Building for Linux based OS ...'
    go build -o ./releases/swap-linux
     ;;

   CYGWIN*|MINGW32*|MSYS*|MINGW64*)
      echo 'Building for Windows ...'
      go build -o ./releases/swap.exe
      # GOOS=windows GOARCH=386 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -o ./releases/swap32.exe -ldflags "-linkmode external -extldflags -static"
     ;;

   # Add here more strings to compare
   # See correspondence table at the bottom of this answer

   *)
     echo 'other OS' 
     ;;
esac
packr clean
rm -rf ui
# rm -rf ui