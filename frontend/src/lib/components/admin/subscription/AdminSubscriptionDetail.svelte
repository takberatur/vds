<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '@/components/ui/badge';

	let {
		open = $bindable(false),
		data,
		onclose
	}: {
		open?: boolean;
		data?: Subscription | null;
		onclose?: () => void;
	} = $props();

	const handleClose = () => {
		open = false;
		onclose?.();
	};
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="min-w-min">
		<Card.Root class="-my-4 w-full max-w-sm">
			<Card.Header>
				<Card.Title>Subscription detail</Card.Title>
				<Card.Description></Card.Description>
			</Card.Header>
			<div class="grid w-full grid-cols-2 gap-2 lg:gap-5">
				<div
					class="col-span-1 gap-2 space-y-2 rounded-md border border-neutral-200 bg-muted p-2 dark:border-neutral-700"
				>
					<div class="text-start font-medium">Transaction ID :</div>
					<div class="text-start font-medium">Product ID :</div>
					<div class="text-start font-medium">Platform :</div>
					<div class="text-start font-medium">Status :</div>
					<div class="text-start font-medium">Auto Renew :</div>
					<div class="text-start font-medium">Start Time :</div>
					<div class="text-start font-medium">End Time :</div>
				</div>
				<div
					class="col-span-1 gap-2 space-y-2 rounded-md border border-neutral-200 bg-muted p-2 dark:border-neutral-700"
				>
					<div class="text-start font-medium">{data?.original_transaction_id}</div>
					<div class="text-start font-medium">{data?.product_id}</div>
					<div class="text-start font-medium">{data?.platform}</div>
					<Badge
						class="text-start font-medium"
						variant={data?.status === 'active'
							? 'default'
							: data?.status === 'expired'
								? 'destructive'
								: 'secondary'}
					>
						{data?.status}
					</Badge>
					<Badge
						class="text-start font-medium"
						variant={data?.auto_renew ? 'default' : 'destructive'}
					>
						{data?.auto_renew ? 'Auto Renew' : 'One Time'}
					</Badge>
					<div class="text-start font-medium">
						{data?.start_time ? new Date(data?.start_time).toLocaleDateString('en-US') : 'N/A'}
					</div>
					<div class="text-start font-medium">
						{data?.end_time ? new Date(data?.end_time).toLocaleDateString('en-US') : 'N/A'}
					</div>
				</div>
			</div>
			<Card.Footer class="flex-col gap-2">
				<Button type="button" class="w-full" onclick={handleClose}>Close</Button>
			</Card.Footer>
		</Card.Root>
	</Dialog.Content>
</Dialog.Root>
