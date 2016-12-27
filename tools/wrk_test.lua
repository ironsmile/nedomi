
--
-- ./wrk -c 100 -d 1s -H 'Host: ggs.com' http://127.0.0.1 -s scripts/wrk_test.lua
--

math.randomseed(os.time())

media_files_weights = {
    ["/Day2/Sofia/07.mp4"] = 1,
    ["/Day2/Sofia/08.mp4"] = 1,
    ["/Day2/Sofia/02.mp4"] = 1,
    ["/Day2/Sofia/06.mp4"] = 1,
    ["/Day2/Sofia/05.mp4"] = 1,
    ["/Day2/Sofia/01.mp4"] = 1,
    ["/Day2/Sofia/04.mp4"] = 1,
    ["/Day2/Sofia/03.mp4"] = 1,
    ["/Day2/G1/02.mp4"] = 1,
    ["/Day2/G1/01.mp4"] = 1,
    ["/Day2/G1/04.mp4"] = 1,
    ["/Day2/G1/03.mp4"] = 1,
    ["/Day2/Varna/07.mp4"] = 1,
    ["/Day2/Varna/02.mp4"] = 1,
    ["/Day2/Varna/06.mp4"] = 1,
    ["/Day2/Varna/05.mp4"] = 1,
    ["/Day2/Varna/01.mp4"] = 1,
    ["/Day2/Varna/04.mp4"] = 2,
    ["/Day2/Varna/03.mp4"] = 2,
    ["/Day1/Sofia/07.mp4"] = 2,
    ["/Day1/Sofia/08.mp4"] = 2,
    ["/Day1/Sofia/02.mp4"] = 2,
    ["/Day1/Sofia/06.mp4"] = 2,
    ["/Day1/Sofia/01.mp4"] = 2,
    ["/Day1/Sofia/04.mp4"] = 2,
    ["/Day1/Sofia/03.mp4"] = 2,
    ["/Day1/G1/02.mp4"] = 3,
    ["/Day1/G1/05.mp4"] = 3,
    ["/Day1/G1/01.mp4"] = 3,
    ["/Day1/G1/04.mp4"] = 3,
    ["/Day1/G1/03.mp4"] = 3,
    ["/Day1/Varna/02.mp4"] = 3,
    ["/Day1/Varna/06.mp4"] = 6,
    ["/Day1/Varna/05.mp4"] = 6,
    ["/Day1/Varna/01.mp4"] = 6,
    ["/Day1/Varna/04.mp4"] = 6,
    ["/Day1/Varna/03.mp4"] = 30,
}

media_files_sizes = {
    ["/filenames"] = 3723,
    ["/Day2/Sofia/07.mp4"] = 310551051,
    ["/Day2/Sofia/08.mp4"] = 209969199,
    ["/Day2/Sofia/02.mp4"] = 348292552,
    ["/Day2/Sofia/06.mp4"] = 382129121,
    ["/Day2/Sofia/05.mp4"] = 439716339,
    ["/Day2/Sofia/01.mp4"] = 394205859,
    ["/Day2/Sofia/04.mp4"] = 286257845,
    ["/Day2/Sofia/03.mp4"] = 372078653,
    ["/Day2/G1/02.mp4"] = 240102115,
    ["/Day2/G1/01.mp4"] = 434461809,
    ["/Day2/G1/04.mp4"] = 418751127,
    ["/Day2/G1/03.mp4"] = 192027551,
    ["/Day2/Varna/07.mp4"] = 362739369,
    ["/Day2/Varna/02.mp4"] = 448981261,
    ["/Day2/Varna/06.mp4"] = 403310150,
    ["/Day2/Varna/05.mp4"] = 460597862,
    ["/Day2/Varna/01.mp4"] = 477849540,
    ["/Day2/Varna/04.mp4"] = 409649259,
    ["/Day2/Varna/03.mp4"] = 321281697,
    ["/index.html"] = 46,
    ["/Day1/Sofia/07.mp4"] = 321781530,
    ["/Day1/Sofia/08.mp4"] = 529576811,
    ["/Day1/Sofia/02.mp4"] = 360469568,
    ["/Day1/Sofia/06.mp4"] = 351310860,
    ["/Day1/Sofia/01.mp4"] = 133265421,
    ["/Day1/Sofia/04.mp4"] = 274970654,
    ["/Day1/Sofia/03.mp4"] = 89962404,
    ["/Day1/G1/02.mp4"] = 149639436,
    ["/Day1/G1/05.mp4"] = 407272649,
    ["/Day1/G1/01.mp4"] = 212734731,
    ["/Day1/G1/04.mp4"] = 356334948,
    ["/Day1/G1/03.mp4"] = 449729900,
    ["/Day1/Varna/02.mp4"] = 397133655,
    ["/Day1/Varna/06.mp4"] = 435394318,
    ["/Day1/Varna/05.mp4"] = 243276331,
    ["/Day1/Varna/01.mp4"] = 358350481,
    ["/Day1/Varna/04.mp4"] = 223466939,
    ["/Day1/Varna/03.mp4"] = 325839813,
}

media_files = {}
media_files_size = 0

BEGINNING = 0 -- Range: bytes=0-<number>
FIRST = 1 -- Range: bytes=<small-number>-number
MIDDLE = 2 -- Rnage: bytes=<middle-number>-number
LAST = 3 -- Range: bytes=<high-number>-<number-possibly-end>

-- Represents the likelihood of request starting its Range from this position
range_starts = {

    -- begging: 50%
    BEGINNING,
    BEGINNING,
    BEGINNING,
    BEGINNING,
    BEGINNING,
    BEGINNING,

    -- First 30m bytes: 25%
    FIRST,
    FIRST,
    FIRST,

    -- In the middle of the file: 8%
    MIDDLE,

    -- Last 30m bytes: 16%
    LAST,
    LAST,
}

range_starts_size = 0

init = function(args)
    wrk.init(args)

    for file, file_weight in pairs(media_files_weights) do
        for _= 0, file_weight do
            table.insert(media_files, file)
        end
    end

    media_files_size = table.getn(media_files)
    range_starts_size = table.getn(range_starts)
end

request = function()    
    
    local file_path = media_files[math.random(1, media_files_size)]
    local file_size = media_files_sizes[file_path]

    -- from 100k to 10m
    local byte_range_size = math.random(100000, 10000000)

    local start_type = range_starts[math.random(1, range_starts_size)]
    local range_start = 0

    if start_type == BEGINNING then
        range_start = 0
    elseif start_type == FIRST then
        range_start = math.random(1, 10000000)
    elseif start_type == LAST then
        range_start = file_size - math.random(1, 10000000)
    else
        range_start = math.random(1, file_size-1)
    end
    
    if range_start < 0 then
        range_start = 0
    end

    range_end = range_start + byte_range_size

    if range_end > file_size then
        range_end = file_size
    end

    wrk.headers["Range"] = string.format("bytes=%d-%d", range_start, range_end)
    wrk.headers["Accept-Encoding"] = "identity;q=1, *;q=0"
    wrk.headers["Accept"] = "*/*"

    return wrk.format(nil, file_path)
end


done = function(summary, latency, requests)
    io.write("------------------------------\n")
    for _, p in pairs({ 50, 90, 99, 99.999 }) do
        n = latency:percentile(p)
        io.write(string.format("%g%%\t%d\n", p, n))
    end
end
