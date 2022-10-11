t1 = tab.new(TEST.url("/download"))
t2 = tab.new(TEST.url("/download"))

download_count = 0
t1:onDownloaded(function(filepath, bytes)
    assert(TEST.storage("data.txt") == filepath)
    assert(bytes == 14)
    download_count = download_count + 1
end)

t2:onDownloaded(function(filepath, bytes)
    error("should not reach here")
end)

t1("a")
    :click()
    :click()

while download_count < 2 do
    time.sleep(5 * time.millisecond)
end

assert(download_count == 2)
assert(io.popen("dir " .. TEST.storage()):read() == "data.txt")

assert(io.input(TEST.storage("data.txt")):read() == "this is a data")