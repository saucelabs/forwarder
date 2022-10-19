// PAC script with getter that changes element kind.

function FindProxyForURL(url, host) {
  let arr = [];
  arr[1000] = 0x1234;

  arr.__defineGetter__(256, function () {
    delete arr[256];
    arr.unshift(1.1);
  });

  let results = Object.entries(arr);
  let str = results.toString();
  return "DIRECT";
}
