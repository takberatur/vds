package com.agcforge.videodownloader.data.model

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
@Parcelize
data class InAppProduct(
    @SerializedName("id") val id: String,
    @SerializedName("app_id") val appId: String,
    @SerializedName("product_id") val productId: String? = null,
    @SerializedName("product_type") val productType: String? = null,
    @SerializedName("sku_code") val skuCode: String? = null,
    @SerializedName("title") val title: String? = null,
    @SerializedName("description") val description: String? = null,
    @SerializedName("price") val price: Double? = null,
    @SerializedName("currency") val currency: String? = null,
    @SerializedName("billing_period") val billingPeriod: String? = null,
    @SerializedName("trial_period_days") val trialPeriodDays: Int? = null,
    @SerializedName("is_active") val isActive: Boolean,
    @SerializedName("is_featured") val isFeatured: Boolean,
    @SerializedName("sort_order") val sortOrder: Int,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String
) : Parcelable
