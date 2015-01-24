
math.randomseed(os.time())

media_files = {
    "/file_1",
    "/file_2",
}

media_files_size = table.getn(media_files)

weighted_random = function(from, to, weighted_for_the_first)
    if not weighted_for_the_first then
        return math.random(from, to)
    end

    return math.random(from, to)
end

request = function()

    req = wrk.format(nil, media_files[weighted_random(1, media_files_size,
            media_files_size/5)])
    wrk.headers["Range"] = string.format("bytes=%d-%d", weighted_random(0, 100),
            weighted_random(100, 10000))

    io.write(string.format("%s\n", req))

    return req
end

done = function(summary, latency, requests)
    io.write("------------------------------\n")
    for _, p in pairs({ 50, 90, 99, 99.999 }) do
        n = latency:percentile(p)
        io.write(string.format("%g%%,%d\n", p, n))
    end
end
