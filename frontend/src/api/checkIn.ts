import request from '@/utils/request'

export function checkInByVoucher(activityId: number, voucher: string) {
  return request.post('/check-ins', { voucher }, { params: { activity_id: activityId } })
}

export function checkInByScan(activityId: number, qrContent: string) {
  return request.post('/check-ins', { qr_content: qrContent }, { params: { activity_id: activityId } })
}

export function listCheckInRecords(activityId: number) {
  return request.get('/check-ins', { params: { activity_id: activityId } })
}
