mkdir build
cd build
git clone --depth=1 https://github.com/alxarno/swap-ui
cd swap-ui
npm install
./node_modules/.bin/webpack
cp -r dist/ ../../ui/
cd ../../
rm -rf build
packr
mkdir releases
case "$(uname -s)" in

   Darwin)
     echo 'Mac OS X'
     ;;

   Linux)
    go build -o ./releases/swap-linux
     ;;

   CYGWIN*|MINGW32*|MSYS*|MINGW64*)
      go build -o ./releases/swap.exe
     ;;

   # Add here more strings to compare
   # See correspondence table at the bottom of this answer

   *)
     echo 'other OS' 
     ;;
esac
packr clean
rm -rf ui