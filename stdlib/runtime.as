static spawn runtime = new (class {
  public get args() {
    return new Array(...#_os_args())
  }
});
