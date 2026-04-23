# `src` Developer Guide

## Muc tieu
- Giam thoi gian onboard cho dev moi.
- Giu code de doc, de tim, de doi.
- Dung mot kieu router duy nhat: Next.js App Router.

## Cau truc thu muc
- `app/`: Next App Router (`layout.tsx`, `page.tsx`, route segments).
- `screens/`: UI cua tung man hinh theo business flow.
- `components/`: UI tai su dung.
- `api/`: client goi API (auth/session).
- `lib/`: utility, route constants, domain helpers.
- `providers/`: React context/provider toan app.
- `i18n/`: message dictionary.
- `types/`: type dung chung.
- `generated/`: file sinh tu script.

## Quy uoc quan trong
1. Khong hardcode URL route trong screen/component.
2. Dung route constants tai `lib/routes.ts`.
3. Route file trong `app/**/page.tsx` chi nen map vao `screens/*`.
4. Component co dung hook/browser API phai co `"use client"`.
5. Bien moi truong uu tien `NEXT_PUBLIC_*` (co the giu fallback cu khi can migrate).

## Pattern tao route moi
1. Them route key trong `lib/routes.ts`.
2. Tao screen moi trong `screens/`.
3. Tao `app/<segment>/page.tsx` va render screen.
4. Neu can link/redirect, import `ROUTES` thay vi ghi chuoi.

## Naming
- `*-page.tsx`: screen cap route.
- `*-provider.tsx`: context provider.
- `*.api.ts`: API client module.
- `components/ui/*`: primitive UI.
- `components/<domain>/*`: component theo nghiep vu.
