import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const BASE_URL = 'http://localhost:5050/api/v1';
let accessToken = '';
const uniqueId = uuidv4();
const username = `Max_${uniqueId}`;
const email = `max_${uniqueId}@mail.com`;

export let options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 0 },
    ],
};

export default function () {

    // Регистрация пользователя
    let registerRes = http.post(`${BASE_URL}/register`, JSON.stringify({
        username: username,
        password: '1234',
        email: email
    }), { headers: { 'Content-Type': 'application/json' } });

    check(registerRes, { 'Register status 201': (r) => r.status === 201 });
    sleep(1);

    // Логин и получение токена
    let loginRes = http.post(`${BASE_URL}/login`, JSON.stringify({
        username: username,
        password: '1234'
    }), { headers: { 'Content-Type': 'application/json' } });

    check(loginRes, { 'Login status 200': (r) => r.status === 200 });

    accessToken = loginRes.json('token');

    // Проверка получения курсов валют
    let ratesRes = http.get(`${BASE_URL}/exchange/rates`);
    check(ratesRes, { 'Rates status 200': (r) => r.status === 200 });
    sleep(1);

    // Депозит
    let depositRes = http.post(`${BASE_URL}/wallet/deposit`, JSON.stringify({
        amount: 100,
        currency: 'EUR'
    }), { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${accessToken}` } });

    check(depositRes, { 'Deposit status 200': (r) => r.status === 200 });
    sleep(1);

    // Вывод средств
    let withdrawRes = http.post(`${BASE_URL}/wallet/withdraw`, JSON.stringify({
        amount: 10,
        currency: 'EUR'
    }), { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${accessToken}` } });

    check(withdrawRes, { 'Withdraw status 200': (r) => r.status === 200 });
    sleep(1);

    // Обмен валют
    let exchangeRes = http.post(`${BASE_URL}/exchange`, JSON.stringify({
        from_currency: 'EUR',
        to_currency: 'USD',
        amount: 10
    }), { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${accessToken}` } });

    check(exchangeRes, { 'Exchange status 200': (r) => r.status === 200 });
    sleep(1);

    // Получение баланса
    let balanceRes = http.get(`${BASE_URL}/balance`, { headers: { 'Authorization': `Bearer ${accessToken}` } });
    check(balanceRes, { 'Balance status 200': (r) => r.status === 200 });
}