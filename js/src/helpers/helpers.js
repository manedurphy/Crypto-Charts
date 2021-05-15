export function handleDateConversion(data) {
    data.forEach((item) => {
        const date = new Date(item.time * 1000);
        item.time = `${date.getMonth()}/${date.getDate()}`;
    });

    return data;
}

export function getUrl(current) {
    let url;
    if (current === 'bar') {
        url = process.env.NODE_ENV === 'development' ? 'http://localhost:8081/api/crypto' : '/api/crypto';
    }

    if (current !== 'bar') {
        url =
            process.env.NODE_ENV === 'development'
                ? `http://localhost:8081/api/crypto/monthly/${current}`
                : `/api/crypto/monthly/${current}`;
    }
    return url;
}
