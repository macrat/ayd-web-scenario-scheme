t = tab.new(os.getenv("TEST_URL"))

print("It's working!")

if os.getenv("TEST_EXTRA") then
    print.extra("msg", os.getenv("TEST_EXTRA"))
end

if os.getenv("TEST_STATUS") then
    print.status(os.getenv("TEST_STATUS"))
end

assert.eq(t("#greeting .target").text, os.getenv("TEST_TEXT"))

if os.getenv("TEST_ERROR") then
    error(os.getenv("TEST_ERROR"))
end