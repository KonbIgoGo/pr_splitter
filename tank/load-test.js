import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    vus: 10,
    duration: '1m',
};

const BASE_URL = 'http://localhost:8080';

let globalUserIds = [];
let globalPrIds = [];
let globalTeamNames = [];

export default function () {
    const teamAddUrl = `${BASE_URL}/team/add`;
    const teamGetUrl = `${BASE_URL}/team/get`;
    const setIsActiveUrl = `${BASE_URL}/users/setIsActive`;
    const getReviewUrl = `${BASE_URL}/users/getReview`;
    const prCreateUrl = `${BASE_URL}/pullRequest/create`;
    const prMergeUrl = `${BASE_URL}/pullRequest/merge`;
    const prReassignUrl = `${BASE_URL}/pullRequest/reassign`;

    const teamName = `team-${__VU}-${__ITER}-${randomIntBetween(1000, 9999)}`;
    const members = [];

    for (let i = 0; i < 4; i++) {
        const userId = `u-${__VU}-${__ITER}-${i}-${randomIntBetween(1000, 9999)}`;
        members.push({
            user_id: userId,
            username: `User_${userId}`,
            is_active: true,
        });
        globalUserIds.push(userId);
    }

    const teamPayload = JSON.stringify({
        team_name: teamName,
        members: members,
    });

    const teamRes = http.post(teamAddUrl, teamPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    let authorId = null;

    check(teamRes, {
        'team created or already exists (201/400)': (r) => r.status === 201 || r.status === 400,
    });

    if (teamRes.status === 201) {
        try {
            const resp = teamRes.json();
            if (resp && resp.team && resp.team.members && resp.team.members.length > 0) {
                authorId = resp.team.members[0].user_id;
                globalTeamNames.push(resp.team.team_name);
            } else {
                console.error('Team response missing members:', teamRes.body);
            }
        } catch (e) {
            console.error('Team JSON parse error:', e.message);
        }
    } else {
        if (globalUserIds.length > 0) {
            authorId = randomItem(globalUserIds);
        }
    }

    if (globalTeamNames.length > 0) {
        const targetTeam = randomItem(globalTeamNames);
        const getTeamRes = http.get(`${teamGetUrl}?team_name=${encodeURIComponent(targetTeam)}`);

        check(getTeamRes, {
            'team get (200 or 404)': (r) => r.status === 200 || r.status === 404,
        });
    }

    let prId = `pr-${__VU}-${__ITER}-${randomIntBetween(1000, 9999)}`;
    let createdPr = null;

    if (authorId) {
        const prPayload = JSON.stringify({
            pull_request_id: prId,
            pull_request_name: `Feature-${randomIntBetween(1000, 9999)}`,
            author_id: authorId,
        });

        const prRes = http.post(prCreateUrl, prPayload, {
            headers: { 'Content-Type': 'application/json' },
        });

        check(prRes, {
            'pr create (201/404/409)': (r) =>
                r.status === 201 || r.status === 404 || r.status === 409,
        });

        if (prRes.status === 201) {
            try {
                const resp = prRes.json();
                if (resp && resp.pr && resp.pr.pull_request_id) {
                    createdPr = resp.pr;
                    prId = resp.pr.pull_request_id;
                    globalPrIds.push(prId);
                } else {
                    console.error('PR response missing pr object:', prRes.body);
                }
            } catch (e) {
                console.error('PR JSON parse error:', e.message);
            }
        } else if (prRes.status === 409) {
            if (globalPrIds.length > 0) {
                prId = randomItem(globalPrIds);
            }
        }
    }

    if (createdPr && createdPr.assigned_reviewers && createdPr.assigned_reviewers.length > 0) {
        const oldUserId = createdPr.assigned_reviewers[0];

        const reassignPayload = JSON.stringify({
            pull_request_id: prId,
            old_user_id: oldUserId,
        });

        const reassignRes = http.post(prReassignUrl, reassignPayload, {
            headers: { 'Content-Type': 'application/json' },
        });

        check(reassignRes, {
            'reassign (200/404/409)': (r) =>
                r.status === 200 || r.status === 404 || r.status === 409,
        });
    }

    if (createdPr && createdPr.assigned_reviewers && createdPr.assigned_reviewers.length > 0) {
        const reviewerId = randomItem(createdPr.assigned_reviewers);
        const getReviewRes = http.get(`${getReviewUrl}?user_id=${encodeURIComponent(reviewerId)}`);

        check(getReviewRes, {
            'getReview (200)': (r) => r.status === 200,
        });
    } else if (globalUserIds.length > 0) {
        const reviewerId = randomItem(globalUserIds);
        const getReviewRes = http.get(`${getReviewUrl}?user_id=${encodeURIComponent(reviewerId)}`);

        check(getReviewRes, {
            'getReview fallback (200)': (r) => r.status === 200,
        });
    }

    if (prId) {
        const mergePayload = JSON.stringify({
            pull_request_id: prId,
        });

        const mergeRes = http.post(prMergeUrl, mergePayload, {
            headers: { 'Content-Type': 'application/json' },
        });

        check(mergeRes, {
            'merge (200/404)': (r) => r.status === 200 || r.status === 404,
        });
    }

    if (globalUserIds.length > 0) {
        const targetUser = randomItem(globalUserIds);
        const isActive = Math.random() > 0.5;

        const setIsActivePayload = JSON.stringify({
            user_id: targetUser,
            is_active: isActive,
        });

        const setIsActiveRes = http.post(setIsActiveUrl, setIsActivePayload, {
            headers: { 'Content-Type': 'application/json' },
        });

        check(setIsActiveRes, {
            'setIsActive (200/404)': (r) => r.status === 200 || r.status === 404,
        });
    }

    sleep(1);
}
